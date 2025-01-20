package jq

import (
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/internal/text"
	"github.com/itchyny/gojq"
)

// WithTemplateFunctions adds some functions from the template package including:
//   - timeago: parses RFC3339 date-times and return relative time e.g., "5 minutes ago".
//   - timefmt: parses RFC3339 date-times,and formats according to layout argument documented at https://pkg.go.dev/time#Layout.
func WithTemplateFunctions() EvaluateOption {
	return func(opts *evaluateOptions) {
		now := time.Now()

		opts.compilerOptions = append(
			opts.compilerOptions,
			gojq.WithFunction("timeago", 0, 0, timeAgoJqFunc(now)),
		)

		opts.compilerOptions = append(
			opts.compilerOptions,
			gojq.WithFunction("timefmt", 1, 1, timeFmtJq),
		)
	}
}

func timeAgoJqFunc(now time.Time) func(v any, _ []any) any {
	return func(v any, _ []any) any {
		if input, ok := v.(string); ok {
			if t, err := text.TimeAgoFunc(now, input); err != nil {
				return cannotFormatError(v, err)
			} else {
				return t
			}
		}

		return notStringError(v)
	}
}

func timeFmtJq(v any, vs []any) any {
	var input, format string
	var ok bool

	if input, ok = v.(string); !ok {
		return notStringError(v)
	}

	if len(vs) != 1 {
		return fmt.Errorf("timefmt requires time format argument")
	}

	if format, ok = vs[0].(string); !ok {
		return notStringError(v)
	}

	if t, err := text.TimeFormatFunc(format, input); err != nil {
		return cannotFormatError(v, err)
	} else {
		return t
	}
}

type valueError struct {
	error
	value any
}

func notStringError(v any) gojq.ValueError {
	return valueError{
		error: fmt.Errorf("%v is not a string", v),
		value: v,
	}
}

func cannotFormatError(v any, err error) gojq.ValueError {
	return valueError{
		error: fmt.Errorf("cannot format %v, %w", v, err),
		value: v,
	}
}

func (v valueError) Value() any {
	return v.value
}
