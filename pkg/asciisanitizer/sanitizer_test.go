package asciisanitizer

import (
	"bytes"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/transform"
)

func TestSanitizerTransform(t *testing.T) {
	tests := []struct {
		name  string
		json  bool
		input string
		want  string
	}{
		{
			name:  "No control characters",
			input: "The quick brown fox jumped over the lazy dog",
			want:  "The quick brown fox jumped over the lazy dog",
		},
		{
			name:  "JSON sanitization maintains valid JSON",
			json:  true,
			input: `\u001B \\u001B \\\u001B \\\\u001B`,
			want:  `^[ \\^[ \\^[ \\\\^[`,
		},
		{
			name:  "JSON C0 control character",
			json:  true,
			input: `0\u0000`,
			want:  "0^@",
		},
		{
			name: "JSON C0 control characters",
			json: true,
			input: `0\u0000 1\u0001 2\u0002 3\u0003 4\u0004 5\u0005 6\u0006 7\u0007 8\u0008 9\u0009 ` +
				`A\u000a B\u000b C\u000c D\u000d E\u000e F\u000f ` +
				`10\u0010 11\u0011 12\u0012 13\u0013 14\u0014 15\u0015 16\u0016 17\u0017 18\u0018 19\u0019 ` +
				`1A\u001a 1B\u001b 1C\u001c 1D\u001d 1E\u001e 1F\u001f`,
			want: `0^@ 1^A 2^B 3^C 4^D 5^E 6^F 7^G 8^H 9\u0009 ` +
				`A\u000a B\u000b C^L D\u000d E^N F^O ` +
				`10^P 11^Q 12^R 13^S 14^T 15^U 16^V 17^W 18^X 19^Y ` +
				`1A^Z 1B^[ 1C^\\ 1D^] 1E^^ 1F^_`,
		},
		{
			name: "JSON C1 control characters",
			json: true,
			input: `80\u0080 81\u0081 82\u0082 83\u0083 84\u0084 85\u0085 86\u0086 87\u0087 88\u0088 89\u0089 ` +
				`8A\u008a 8B\u008b 8C\u008c 8D\u008d 8E\u008e 8F\u008f ` +
				`90\u0090 91\u0091 92\u0092 93\u0093 94\u0094 95\u0095 96\u0096 97\u0097 98\u0098 99\u0099 ` +
				`9A\u009a 9B\u009b 9C\u009c 9D\u009d 9E\u009e 9F\u009f`,
			want: `80^@ 81^A 82^B 83^C 84^D 85^E 86^F 87^G 88^H 89^I ` +
				`8A^J 8B^K 8C^L 8D^M 8E^N 8F^O ` +
				`90^P 91^Q 92^R 93^S 94^T 95^U 96^V 97^W 98^X 99^Y ` +
				`9A^Z 9B^[ 9C^\\ 9D^] 9E^^ 9F^_`,
		},
		{
			name:  "C0 control character",
			input: "0\x00",
			want:  "0^@",
		},
		{
			name: "C0 control characters",
			input: "0\x00 1\x01 2\x02 3\x03 4\x04 5\x05 6\x06 7\x07 8\x08 9\x09 " +
				"A\x0A B\x0B C\x0C D\x0D E\x0E F\x0F " +
				"10\x10 11\x11 12\x12 13\x13 14\x14 15\x15 16\x16 17\x17 18\x18 19\x19 " +
				"1A\x1A 1B\x1B 1C\x1C 1D\x1D 1E\x1E 1F\x1F",
			want: "0^@ 1^A 2^B 3^C 4^D 5^E 6^F 7^G 8^H 9\t " +
				"A\n B\v C^L D\r E^N F^O " +
				"10^P 11^Q 12^R 13^S 14^T 15^U 16^V 17^W 18^X 19^Y " +
				"1A^Z 1B^[ 1C^\\\\ 1D^] 1E^^ 1F^_",
		},
		{
			name:  "C1 control character",
			input: "80\xC2\x80",
			want:  "80^@",
		},
		{
			name: "C1 control characters",
			input: "80\xC2\x80 81\xC2\x81 82\xC2\x82 83\xC2\x83 84\xC2\x84 85\xC2\x85 86\xC2\x86 87\xC2\x87 88\xC2\x88 89\xC2\x89 " +
				"8A\xC2\x8A 8B\xC2\x8B 8C\xC2\x8C 8D\xC2\x8D 8E\xC2\x8E 8F\xC2\x8F " +
				"90\xC2\x90 91\xC2\x91 92\xC2\x92 93\xC2\x93 94\xC2\x94 95\xC2\x95 96\xC2\x96 97\xC2\x97 98\xC2\x98 99\xC2\x99 " +
				"9A\xC2\x9A 9B\xC2\x9B 9C\xC2\x9C 9D\xC2\x9D 9E\xC2\x9E 9F\xC2\x9F",
			want: "80^@ 81^A 82^B 83^C 84^D 85^E 86^F 87^G 88^H 89^I " +
				"8A^J 8B^K 8C^L 8D^M 8E^N 8F^O " +
				"90^P 91^Q 92^R 93^S 94^T 95^U 96^V 97^W 98^X 99^Y " +
				"9A^Z 9B^[ 9C^\\\\ 9D^] 9E^^ 9F^_",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitizer := &Sanitizer{JSON: tt.json}
			reader := bytes.NewReader([]byte(tt.input))
			transformReader := transform.NewReader(reader, sanitizer)
			err := iotest.TestReader(transformReader, []byte(tt.want))
			require.NoError(t, err)
		})
	}
}
