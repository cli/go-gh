package jsonpretty

import (
	"bytes"
	"io"
	"testing"
)

func TestWrite(t *testing.T) {
	type args struct {
		r        io.Reader
		indent   string
		colorize bool
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "blank",
			args: args{
				r:        bytes.NewBufferString(``),
				indent:   "",
				colorize: true,
			},
			wantW:   "",
			wantErr: false,
		},
		{
			name: "empty object",
			args: args{
				r:        bytes.NewBufferString(`{}`),
				indent:   "",
				colorize: true,
			},
			wantW:   "\x1b[1;38m{\x1b[m\x1b[1;38m}\x1b[m\n",
			wantErr: false,
		},
		{
			name: "nested object",
			args: args{
				r:        bytes.NewBufferString(`{"hash":{"a":1,"b":2},"array":[3,4]}`),
				indent:   "\t",
				colorize: true,
			},
			wantW: "\x1b[1;38m{\x1b[m\n\t\x1b[1;34m\"hash\"\x1b[m\x1b[1;38m:\x1b[m " +
				"\x1b[1;38m{\x1b[m\n\t\t\x1b[1;34m\"a\"\x1b[m\x1b[1;38m:\x1b[m 1\x1b[1;38m,\x1b[m\n\t\t\x1b[1;34m\"b\"\x1b[m\x1b[1;38m:\x1b[m 2\n\t\x1b[1;38m}\x1b[m\x1b[1;38m,\x1b[m" +
				"\n\t\x1b[1;34m\"array\"\x1b[m\x1b[1;38m:\x1b[m \x1b[1;38m[\x1b[m\n\t\t3\x1b[1;38m,\x1b[m\n\t\t4\n\t\x1b[1;38m]\x1b[m\n\x1b[1;38m}\x1b[m\n",
			wantErr: false,
		},
		{
			name: "no color",
			args: args{
				r:        bytes.NewBufferString(`{"hash":{"a":1,"b":2},"array":[3,4]}`),
				indent:   "\t",
				colorize: false,
			},
			wantW:   "{\n\t\"hash\": {\n\t\t\"a\": 1,\n\t\t\"b\": 2\n\t},\n\t\"array\": [\n\t\t3,\n\t\t4\n\t]\n}\n",
			wantErr: false,
		},
		{
			name: "string",
			args: args{
				r:        bytes.NewBufferString(`"foo"`),
				indent:   "",
				colorize: true,
			},
			wantW:   "\x1b[32m\"foo\"\x1b[m\n",
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				r:        bytes.NewBufferString(`{{`),
				indent:   "",
				colorize: true,
			},
			wantW:   "\x1b[1;38m{\x1b[m\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := Format(w, tt.args.r, tt.args.indent, tt.args.colorize); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if w.String() != tt.wantW {
				t.Errorf("got: %q, want: %q", w.String(), tt.wantW)
			}
		})
	}
}
