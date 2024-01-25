package jsonmerge

import (
	"bytes"
	"encoding/json"
	"io"

	"dario.cat/mergo"
)

type objectMerger struct {
	io.Writer
	dst map[string]interface{}
}

// NewObjectMerger creates a Merger for JSON objects.
func NewObjectMerger(w io.Writer) Merger {
	return &objectMerger{
		Writer: w,
		dst:    make(map[string]interface{}),
	}
}

func (merger *objectMerger) NewPage(r io.Reader, isLastPage bool) io.ReadCloser {
	return &objectMergerPage{
		merger: merger,
		Reader: r,
	}
}

func (merger *objectMerger) Close() error {
	// Marshal to JSON and write to output.
	buf, err := json.Marshal(merger.dst)
	if err != nil {
		return err
	}

	_, err = merger.Writer.Write(buf)
	return err
}

type objectMergerPage struct {
	merger *objectMerger

	io.Reader
	buffer bytes.Buffer
}

func (page *objectMergerPage) Read(p []byte) (int, error) {
	// Read into a temporary buffer to be merged and written later.
	p = make([]byte, len(p))
	n, err := page.Reader.Read(p)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}

	_, err = page.buffer.Write(p[:n])
	return 0, err
}

func (page *objectMergerPage) Close() error {
	var src map[string]interface{}

	err := json.Unmarshal(page.buffer.Bytes(), &src)
	if err != nil {
		return err
	}

	return mergo.Merge(&page.merger.dst, src, mergo.WithAppendSlice, mergo.WithOverride)
}
