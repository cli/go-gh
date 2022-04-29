package transportmock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Responder func(req *http.Request) (*http.Response, error)

func HTTPResponse(status int, header *map[string][]string, body interface{}, cb func(*http.Request)) Responder {
	return func(req *http.Request) (*http.Response, error) {
		var b io.Reader
		if s, ok := body.(string); ok {
			b = bytes.NewBufferString(s)
		} else {
			s, _ := json.Marshal(body)
			b = bytes.NewBuffer(s)
		}
		if cb != nil {
			cb(req)
		}
		if header == nil {
			header = &map[string][]string{}
		}
		return httpResponse(status, *header, b, req), nil
	}
}

func RESTResponse(body interface{}, cb func(map[string]interface{})) Responder {
	return func(req *http.Request) (*http.Response, error) {
		var b io.Reader
		if s, ok := body.(string); ok {
			b = bytes.NewBufferString(s)
		} else {
			s, _ := json.Marshal(body)
			b = bytes.NewBuffer(s)
		}
		bodyData := map[string]interface{}{}
		err := decodeJSONBody(req, &bodyData)
		if err != nil {
			return nil, err
		}
		if cb != nil {
			cb(bodyData)
		}
		return httpResponse(200, map[string][]string{}, b, req), nil
	}
}

func GQLMutation(body interface{}, cb func(map[string]interface{})) Responder {
	return func(req *http.Request) (*http.Response, error) {
		var b io.Reader
		if s, ok := body.(string); ok {
			b = bytes.NewBufferString(s)
		} else {
			s, _ := json.Marshal(body)
			b = bytes.NewBuffer(s)
		}
		var bodyData struct {
			Variables struct {
				Input map[string]interface{}
			}
		}
		err := decodeJSONBody(req, &bodyData)
		if err != nil {
			return nil, err
		}
		if cb != nil {
			cb(bodyData.Variables.Input)
		}
		return httpResponse(200, map[string][]string{}, b, req), nil
	}
}

func GQLQuery(body interface{}, cb func(string, map[string]interface{})) Responder {
	return func(req *http.Request) (*http.Response, error) {
		var b io.Reader
		if s, ok := body.(string); ok {
			b = bytes.NewBufferString(s)
		} else {
			s, _ := json.Marshal(body)
			b = bytes.NewBuffer(s)
		}
		var bodyData struct {
			Query     string
			Variables map[string]interface{}
		}
		err := decodeJSONBody(req, &bodyData)
		if err != nil {
			return nil, err
		}
		if cb != nil {
			cb(bodyData.Query, bodyData.Variables)
		}
		return httpResponse(200, map[string][]string{}, b, req), nil
	}
}

func httpResponse(status int, header map[string][]string, body io.Reader, req *http.Request) *http.Response {
	if _, ok := header["Content-Type"]; !ok {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
	return &http.Response{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(body),
		Request:    req,
	}
}
