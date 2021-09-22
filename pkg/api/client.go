package api

import (
	"io"
	"net/http"
	"time"
)

type ClientOptions struct {
	AuthToken   string
	CacheDir    string
	CacheTTL    time.Duration
	EnableCache bool
	Headers     map[string]string
	Host        string
	Log         io.Writer
	Timeout     time.Duration
	Transport   http.RoundTripper
}

type RESTClient interface {
	Do(method string, path string, body io.Reader, response interface{}) error
	Delete(path string, response interface{}) error
	Get(path string, response interface{}) error
	Patch(path string, body io.Reader, response interface{}) error
	Post(path string, body io.Reader, response interface{}) error
	Put(path string, body io.Reader, response interface{}) error
}

type GQLClient interface {
	Do(query string, variables map[string]interface{}, data interface{}) error
	Mutate(mutation interface{}, input interface{}, variables map[string]interface{}) error
	Query(query interface{}, variables map[string]interface{}) error
}
