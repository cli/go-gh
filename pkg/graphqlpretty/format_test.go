package graphqlpretty

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	var w bytes.Buffer
	err := Format(&w, strings.NewReader(`{query($hello: Int!){repo(name:"keksi\"beksi",first:5){foo,bar}}}`), "  ", false)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if w.String() != "" {
		t.Errorf("got\n%s", w.String())
	}
}
