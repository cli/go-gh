// Package jq facilitates processing of JSON strings using jq expressions.
package jq

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/itchyny/gojq"
)

// Evaluate a jq expression against an input and write it to an output.
func Evaluate(input io.Reader, output io.Writer, expr string) error {
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

	enc := json.NewEncoder(output)

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
			_, err = fmt.Fprint(output, "[]\n")
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
