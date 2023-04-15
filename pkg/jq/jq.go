// Package jq facilitates processing of JSON strings using jq expressions.
package jq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/jsonpretty"
	"github.com/itchyny/gojq"
)

// Evaluate a jq expression against an input and write it to an output.
// Any top-level scalar values produced by the jq expression are written out
// directly, as raw values and not as JSON scalars, similar to how jq --raw
// works.
func Evaluate(input io.Reader, output io.Writer, expr string) error {
	return EvaluateFormatted(input, output, expr, "", false)
}

// Evaluate a jq expression against an input and write it to an output,
// optionally with indentation and colorization.  Any top-level scalar values
// produced by the jq expression are written out directly, as raw values and not
// as JSON scalars, similar to how jq --raw works.
func EvaluateFormatted(input io.Reader, output io.Writer, expr string, indent string, colorize bool) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return err
	}

	code, err := gojq.Compile(
		query,
		gojq.WithEnvironLoader(func() []string {
			return os.Environ()
		}))
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
