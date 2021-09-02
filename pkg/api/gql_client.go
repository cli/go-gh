package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shurcooL/githubv4"
)

type GQLClient interface {
	Do(query string, variables map[string]interface{}, data interface{}) error
	Mutate(mutation interface{}, input interface{}, variables map[string]interface{}) error
	Query(query interface{}, variables map[string]interface{}) error
}

type gqlClient struct {
	client     *githubv4.Client
	host       string
	httpClient *http.Client
}

func NewGQLClient(host string, opts ClientOptions) GQLClient {
	httpClient := newHTTPClient(opts)

	var client *githubv4.Client
	if isEnterprise(host) {
		host = fmt.Sprintf("https://%s/api/graphql", host)
		client = githubv4.NewEnterpriseClient(host, &httpClient)
	} else {
		host = "https://api.github.com/graphql"
		client = githubv4.NewClient(&httpClient)
	}

	return gqlClient{
		client:     client,
		host:       host,
		httpClient: &httpClient,
	}
}

// Do executes a single GraphQL query request and parses the response
func (c gqlClient) Do(query string, variables map[string]interface{}, data interface{}) error {
	reqBody, err := json.Marshal(map[string]interface{}{"query": query, "variables": variables})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.host, bytes.NewBuffer(reqBody))
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
		return handleHTTPError(resp)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	gr := &gqlResponse{Data: data}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		return err
	}

	if len(gr.Errors) > 0 {
		return &gqlErrorResponse{Errors: gr.Errors}
	}

	return nil
}

// Mutate executes a single GraphQL mutation request,
// with a mutation derived from m, populating the response into it.
// m should be a pointer to struct that corresponds to the GitHub GraphQL schema.
// Provided input will be set as a variable named "input".
func (c gqlClient) Mutate(m interface{}, input interface{}, variables map[string]interface{}) error {
	return c.client.Mutate(context.Background(), m, input, variables)
}

// Query executes a single GraphQL query request,
// with a query derived from q, populating the response into it.
// q should be a pointer to struct that corresponds to the GitHub GraphQL schema.
func (c gqlClient) Query(q interface{}, variables map[string]interface{}) error {
	return c.client.Query(context.Background(), q, variables)
}

type gqlResponse struct {
	Data   interface{}
	Errors []gqlError
}

// gqlError is a single error returned in a GQL response
type gqlError struct {
	Type    string
	Message string
}

// gqlErrorResponse contains errors returned in a GQL response
type gqlErrorResponse struct {
	Errors []gqlError
}

func (gr gqlErrorResponse) Error() string {
	errorMessages := make([]string, 0, len(gr.Errors))
	for _, e := range gr.Errors {
		errorMessages = append(errorMessages, e.Message)
	}
	return fmt.Sprintf("GQL error: %s", strings.Join(errorMessages, "\n"))
}
