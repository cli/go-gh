package api

import (
	"net/http"

	"github.com/cli/go-gh/pkg/api"
)

func NewHTTPClient(opts *api.ClientOptions) *http.Client {
	c := newHTTPClient(opts)
	return &c
}
