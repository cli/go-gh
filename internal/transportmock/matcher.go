package transportmock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type Matcher func(req *http.Request) bool

func REST(method, p string) Matcher {
	return func(req *http.Request) bool {
		if !strings.EqualFold(req.Method, method) {
			return false
		}
		if req.URL.Path != "/"+p {
			return false
		}
		return true
	}
}

func GQL(q string) Matcher {
	re := regexp.MustCompile(q)
	return func(req *http.Request) bool {
		if !strings.EqualFold(req.Method, "POST") {
			return false
		}
		if req.URL.Path != "/graphql" && req.URL.Path != "/api/graphql" {
			return false
		}

		var bodyData struct {
			Query string
		}
		_ = decodeJSONBody(req, &bodyData)
		return re.MatchString(bodyData.Query)
	}
}

func decodeJSONBody(req *http.Request, dest interface{}) error {
	if req.Body == nil {
		return nil
	}
	b, err := readBody(req)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func readBody(req *http.Request) ([]byte, error) {
	bodyCopy := &bytes.Buffer{}
	r := io.TeeReader(req.Body, bodyCopy)
	req.Body = io.NopCloser(bodyCopy)
	return io.ReadAll(r)
}
