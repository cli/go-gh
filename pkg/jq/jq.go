// Package jq facilitates processing of JSON strings using jq expressions.
package jq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/jsonpretty"
	"github.com/itchyny/gojq"
)

// evaluateOptions is passed to an EvaluationOption function.
type evaluateOptions struct {
	compilerOptions []gojq.CompilerOption
}

// EvaluateOption is used to configure the pkg/jq.Evaluate functions.
type EvaluateOption func(*evaluateOptions)

// WithModulePaths sets the jq module lookup paths e.g.,
// "~/.jq", "$ORIGIN/../lib/gh", and "$ORIGIN/../lib".
func WithModulePaths(paths []string) EvaluateOption {
	return func(opts *evaluateOptions) {
		opts.compilerOptions = append(
			opts.compilerOptions,
			gojq.WithModuleLoader(gojq.NewModuleLoader(paths)),
		)
	}
}

// Evaluate a jq expression against an input and write it to an output.
// Any top-level scalar values produced by the jq expression are written out
// directly, as raw values and not as JSON scalars, similar to how jq --raw
// works.
func Evaluate(input io.Reader, output io.Writer, expr string, options ...EvaluateOption) error {
	return EvaluateFormatted(input, output, expr, "", false, options...)
}

// Evaluate a jq expression against an input and write it to an output,
// optionally with indentation and colorization.  Any top-level scalar values
// produced by the jq expression are written out directly, as raw values and not
// as JSON scalars, similar to how jq --raw works.
func EvaluateFormatted(input io.Reader, output io.Writer, expr string, indent string, colorize bool, options ...EvaluateOption) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		var e *gojq.ParseError
		if errors.As(err, &e) {
			str, line, column := getLineColumn(expr, e.Offset-len(e.Token))
			return fmt.Errorf(
				"failed to parse jq expression (line %d, column %d)\n    %s\n    %*c  %w",
				line, column, str, column, '^', err,
			)
		}
		return err
	}

	opts := evaluateOptions{
		// Default compiler options.
		compilerOptions: []gojq.CompilerOption{
			gojq.WithEnvironLoader(func() []string {
				return os.Environ()
			}),
		},
	}
	for _, opt := range options {
		opt(&opts)
	}

	code, err := gojq.Compile(
		query,
		opts.compilerOptions...)
	if err != nil {
		return err
	}

	jsonData, err := io.ReadAll(input)
	if err != nil {
		return err
	}

	var responseData interface{}
	err = json.Unmarshal(jsonData, &responseData)
	if err != nil {
		return err
	}

	enc := prettyEncoder{
		w:        output,
		indent:   indent,
		colorize: colorize,
	}

	iter := code.Run(responseData)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			var e *gojq.HaltError
			if errors.As(err, &e) && e.Value() == nil {
				break
			}
			return err
		}
		if text, e := jsonScalarToString(v); e == nil {
			_, err := fmt.Fprintln(output, text)
			if err != nil {
				return err
			}
		} else {
			if err = enc.Encode(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func jsonScalarToString(input interface{}) (string, error) {
	switch tt := input.(type) {
	case string:
		return tt, nil
	case float64:
		if math.Trunc(tt) == tt {
			return strconv.FormatFloat(tt, 'f', 0, 64), nil
		} else {
			return strconv.FormatFloat(tt, 'f', 2, 64), nil
		}
	case nil:
		return "", nil
	case bool:
		return fmt.Sprintf("%v", tt), nil
	default:
		return "", fmt.Errorf("cannot convert type to string: %v", tt)
	}
}

type prettyEncoder struct {
	w        io.Writer
	indent   string
	colorize bool
}

func (p prettyEncoder) Encode(v any) error {
	var b []byte
	var err error
	if p.indent == "" {
		b, err = json.Marshal(v)
	} else {
		b, err = json.MarshalIndent(v, "", p.indent)
	}
	if err != nil {
		return err
	}
	if !p.colorize {
		if _, err := p.w.Write(b); err != nil {
			return err
		}
		if _, err := p.w.Write([]byte{'\n'}); err != nil {
			return err
		}
		return nil
	}
	return jsonpretty.Format(p.w, bytes.NewReader(b), p.indent, true)
}

func getLineColumn(expr string, offset int) (string, int, int) {
	for line := 1; ; line++ {
		index := strings.Index(expr, "\n")
		if index < 0 {
			return expr, line, offset + 1
		}
		if index >= offset {
			return expr[:index], line, offset + 1
		}
		expr = expr[index+1:]
		offset -= index + 1
	}
}
