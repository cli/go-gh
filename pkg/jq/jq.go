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

	"github.com/cli/go-gh/pkg/jsonpretty"

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

	var enc *json.Encoder
	var buff bytes.Buffer
	if !colorize {
		// write straight to the output
		// we can't use jsonpretty here because it handles indent = ""
		// differently from json.Encoder (delimiters would be put on
		// separate lines)
		enc = json.NewEncoder(output)
		enc.SetIndent("", indent)
	} else {
		// write to a buffer, and and then have jsonpretty format from this
		buff = bytes.Buffer{}
		enc = json.NewEncoder(&buff)
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
		} else if tt, ok := v.([]interface{}); ok && tt == nil {
			if err = jsonpretty.Format(output, bytes.NewBuffer([]byte("[]\n")), indent, colorize); err != nil {
				return err
			}
		} else {
			if err = enc.Encode(v); err != nil {
				return err
			}
			if colorize {
				// the encoder has writter to buff, now format it
				if err = jsonpretty.Format(output, &buff, indent, true); err != nil {
					return err
				}
				buff.Reset()
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
