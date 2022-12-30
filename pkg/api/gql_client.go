package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	graphql "github.com/cli/shurcooL-graphql"
)

// GQLClient wraps methods for the different types of
// API requests that are supported by the server.
type GQLClient struct {
	client     *graphql.Client
	host       string
	httpClient *http.Client
}

func DefaultGQLClient() (*GQLClient, error) {
	return NewGQLClient(ClientOptions{})
}

// GQLClient builds a client to send requests to GitHub GraphQL API endpoints.
// As part of the configuration a hostname, auth token, default set of headers,
// and unix domain socket are resolved from the gh environment configuration.
// These behaviors can be overridden using the opts argument.
func NewGQLClient(opts ClientOptions) (*GQLClient, error) {
	if optionsNeedResolution(opts) {
		var err error
		opts, err = resolveOptions(opts)
		if err != nil {
			return nil, err
		}
	}

	httpClient, err := NewHTTPClient(opts)
	if err != nil {
		return nil, err
	}

	endpoint := gqlEndpoint(opts.Host)

	return &GQLClient{
		client:     graphql.NewClient(endpoint, httpClient),
		host:       endpoint,
		httpClient: httpClient,
	}, nil
}

// DoWithContext executes a GraphQL query request.
// The response is populated into the response argument.
func (c *GQLClient) DoWithContext(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	reqBody, err := json.Marshal(map[string]interface{}{"query": query, "variables": variables})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.host, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		return HandleHTTPError(resp)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	gr := gqlResponse{Data: response}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		return err
	}

	if len(gr.Errors) > 0 {
		return &GQLError{Errors: gr.Errors}
	}

	return nil
}

// Do wraps DoWithContext using context.Background.
func (c *GQLClient) Do(query string, variables map[string]interface{}, response interface{}) error {
	return c.DoWithContext(context.Background(), query, variables, response)
}

// MutateWithContext executes a GraphQL mutation request.
// The mutation string is derived from the mutation argument, and the
// response is populated into it.
// The mutation argument should be a pointer to struct that corresponds
// to the GitHub GraphQL schema.
// Provided input will be set as a variable named input.
func (c *GQLClient) MutateWithContext(ctx context.Context, name string, m interface{}, variables map[string]interface{}) error {
	return c.client.MutateNamed(ctx, name, m, variables)
}

// Mutate wraps MutateWithContext using context.Background.
func (c *GQLClient) Mutate(name string, m interface{}, variables map[string]interface{}) error {
	return c.MutateWithContext(context.Background(), name, m, variables)
}

// QueryWithContext executes a GraphQL query request,
// The query string is derived from the query argument, and the
// response is populated into it.
// The query argument should be a pointer to struct that corresponds
// to the GitHub GraphQL schema.
func (c *GQLClient) QueryWithContext(ctx context.Context, name string, q interface{}, variables map[string]interface{}) error {
	return c.client.QueryNamed(ctx, name, q, variables)
}

// Query wraps QueryWithContext using context.Background.
func (c *GQLClient) Query(name string, q interface{}, variables map[string]interface{}) error {
	return c.QueryWithContext(context.Background(), name, q, variables)
}

type gqlResponse struct {
	Data   interface{}
	Errors []GQLErrorItem
}

func gqlEndpoint(host string) string {
	if isGarage(host) {
		return fmt.Sprintf("https://%s/api/graphql", host)
	}
	host = normalizeHostname(host)
	if isEnterprise(host) {
		return fmt.Sprintf("https://%s/api/graphql", host)
	}
	if strings.EqualFold(host, localhost) {
		return fmt.Sprintf("http://api.%s/graphql", host)
	}
	return fmt.Sprintf("https://api.%s/graphql", host)
}
