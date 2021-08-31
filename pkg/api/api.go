package api

import (
	"io"
	"log"
	"net/http"
	"time"
)

type RESTClient interface {
	Do(method string, path string, body io.Reader, resp interface{}) error
	Delete(path string, resp interface{}) error
	Get(path string, resp interface{}) error
	Patch(path string, resp interface{}) error
	Post(path string, body io.Reader, resp interface{}) error
	Put(path string, body io.Reader, resp interface{}) error
}

type GQLClient interface {
	Mutate(m interface{}, input interface{}, variables map[string]interface{}) error
	Query(q interface{}, variables map[string]interface{}) error
}

//Support headers, verbose logging, and caching
type ClientOptions struct {
	AuthToken string
	Headers   map[string]string
	Host      string
	Logger    *log.Logger
	Timeout   time.Duration
	Transport http.RoundTripper
}

type restClient struct {
	ClientOptions
}

type gqlClient struct {
	ClientOptions
}

func NewRESTClient(opts ClientOptions) RESTClient {
	return restClient{opts}
}

func NewGQLClient(opts ClientOptions) GQLClient {
	return gqlClient{opts}
}

func (c restClient) Do(method string, path string, body io.Reader, resp interface{}) error {
	return nil
}

func (c restClient) Delete(path string, resp interface{}) error {
	return nil
}

func (c restClient) Get(path string, resp interface{}) error {
	return nil
}

func (c restClient) Patch(path string, resp interface{}) error {
	return nil
}

func (c restClient) Post(path string, body io.Reader, resp interface{}) error {
	return nil
}

func (c restClient) Put(path string, body io.Reader, resp interface{}) error {
	return nil
}

func (c gqlClient) Mutate(m interface{}, input interface{}, variables map[string]interface{}) error {
	return nil
}

func (c gqlClient) Query(q interface{}, variables map[string]interface{}) error {
	return nil
}
