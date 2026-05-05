package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGraphQLContainsMutationOrSubscription(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		expected bool
	}{
		{"shorthand selection set", "{ viewer { login } }", false},
		{"shorthand with leading whitespace", "  \n\t{ viewer { login } }", false},
		{"named query", "query Q { viewer { login } }", false},
		{"named query with vars", "query Q($x: Int) { node(id: $x) { id } }", false},
		{"query with default var", "query Q($x: Int = 5) { y }", false},
		{"query with directive", "query Q @retry { y }", false},
		{"named mutation", "mutation M { x { id } }", true},
		{"named subscription", "subscription S { events { id } }", true},
		{"unnamed mutation", "mutation { x { id } }", true},
		{"unnamed subscription", "subscription { x { id } }", true},
		{"mutation with comment header", "# this is a mutation\nmutation M { x }", true},
		{"comment with mutation word but query body", "# this is a mutation\nquery Q { x }", false},
		{"multi-op query then mutation", "query Q { x } mutation M { y }", true},
		{"multi-op fragment then mutation", "fragment F on T { y } mutation M { x }", true},
		{"multi-op query then fragment", "query Q { x ...F } fragment F on T { y }", false},
		{"fragment only", "fragment F on T { y }", false},
		{"field aliased mutation inside selection", "query Q { mutation: viewer { login } }", false},
		{"variable named mutation", "query Q($mutation: String) { x(arg: $mutation) }", false},
		{"argument named mutation", "query Q { x(mutation: 1) { y } }", false},
		{"string literal containing mutation keyword", `query Q { x(arg: "this is a mutation") { y } }`, false},
		{"block string with mutation keyword", "query Q { x(arg: \"\"\"\nmutation\n\"\"\") }", false},
		{"directive named mutation on field", "query Q { x @mutation }", false},
		{"directive named mutation on operation", "query Q @mutation { x }", false},
		{"unknown top-level keyword fails closed", "weird Q { x }", true},
		{"empty document", "", false},
		{"whitespace only", "   \n  \t", false},
		{"BOM then query", "\xEF\xBB\xBFquery Q { x }", false},
		{"BOM then mutation", "\xEF\xBB\xBFmutation M { x }", true},
		{"name prefix containing mutation is not a match", "query mutationOne { x }", false},
		{"unterminated string fails closed", `query Q { x(arg: "unterminated`, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := graphqlContainsMutationOrSubscription(c.query)
			assert.Equal(t, c.expected, got)
		})
	}
}

func TestIsCacheableGraphQLRequest(t *testing.T) {
	makeReq := func(body string) *http.Request {
		req, err := http.NewRequest("POST", "https://api.github.com/graphql", strings.NewReader(body))
		assert.NoError(t, err)
		return req
	}

	t.Run("query body is cacheable", func(t *testing.T) {
		req := makeReq(`{"query":"query Q { viewer { login } }","variables":{}}`)
		assert.True(t, isCacheableGraphQLRequest(req))
	})

	t.Run("mutation body is not cacheable", func(t *testing.T) {
		req := makeReq(`{"query":"mutation M { addStar(input: {starrableId: \"x\"}) { clientMutationId } }"}`)
		assert.False(t, isCacheableGraphQLRequest(req))
	})

	t.Run("subscription body is not cacheable", func(t *testing.T) {
		req := makeReq(`{"query":"subscription S { events { id } }"}`)
		assert.False(t, isCacheableGraphQLRequest(req))
	})

	t.Run("invalid JSON body is not cacheable (fail closed)", func(t *testing.T) {
		req := makeReq(`not json`)
		assert.False(t, isCacheableGraphQLRequest(req))
	})

	t.Run("missing query field is not cacheable (fail closed)", func(t *testing.T) {
		req := makeReq(`{"variables":{}}`)
		assert.False(t, isCacheableGraphQLRequest(req))
	})

	t.Run("nil body is not cacheable", func(t *testing.T) {
		req, err := http.NewRequest("POST", "https://api.github.com/graphql", nil)
		assert.NoError(t, err)
		assert.False(t, isCacheableGraphQLRequest(req))
	})

	t.Run("body remains readable for downstream consumers", func(t *testing.T) {
		body := `{"query":"query Q { viewer { login } }"}`
		req := makeReq(body)
		_ = isCacheableGraphQLRequest(req)

		downstream, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, body, string(downstream), "body must still be fully readable after the cacheability peek")
	})
}

// TestCacheRoundTrip_GraphQLMutation_NotCached verifies end-to-end that a
// GraphQL mutation request goes to the network on every call rather than
// being served from cache.
func TestCacheRoundTrip_GraphQLMutation_NotCached(t *testing.T) {
	counter := 0
	fakeHTTP := tripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			counter++
			body := fmt.Sprintf("response %d", counter)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
			}, nil
		},
	}

	cacheDir := filepath.Join(t.TempDir(), "gh-cli-cache")
	httpClient, err := NewHTTPClient(
		ClientOptions{
			Host:         "github.com",
			AuthToken:    "token",
			Transport:    fakeHTTP,
			EnableCache:  true,
			CacheTTL:     time.Hour,
			CacheDir:     cacheDir,
			LogIgnoreEnv: true,
		},
	)
	assert.NoError(t, err)

	do := func(graphqlBody string) string {
		req, err := http.NewRequest("POST", "https://api.github.com/graphql", strings.NewReader(graphqlBody))
		assert.NoError(t, err)
		res, err := httpClient.Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		return string(body)
	}

	// Query is cacheable: second call returns the same body.
	assert.Equal(t, "response 1", do(`{"query":"query Q { viewer { login } }"}`))
	assert.Equal(t, "response 1", do(`{"query":"query Q { viewer { login } }"}`))

	// Mutation is not cacheable: each call hits the network.
	assert.Equal(t, "response 2", do(`{"query":"mutation M { x { id } }"}`))
	assert.Equal(t, "response 3", do(`{"query":"mutation M { x { id } }"}`))
}
