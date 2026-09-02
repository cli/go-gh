package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/cli/go-gh/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// thirdPartyHost is a host that is neither the canonical API host nor the configured
// api_host, and which must therefore never be sent the auth token.
const thirdPartyHost = "unrelated.example"

type recordedRequest struct {
	method        string
	path          string
	rawQuery      string
	host          string
	authorization string
}

type requestRecorder struct {
	mu       sync.Mutex
	requests []recordedRequest
}

func (recorder *requestRecorder) record(req *http.Request) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	recorder.requests = append(recorder.requests, recordedRequest{
		method:        req.Method,
		path:          req.URL.Path,
		rawQuery:      req.URL.RawQuery,
		host:          req.Host,
		authorization: req.Header.Get("Authorization"),
	})
}

func (recorder *requestRecorder) recordedRequests() []recordedRequest {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return append([]recordedRequest(nil), recorder.requests...)
}

func requireRequest(t *testing.T, recorder *requestRecorder, want recordedRequest) {
	t.Helper()

	require.Contains(t, recorder.recordedRequests(), want)
}

type apiHostTestHarness struct {
	transport          *http.Transport
	githubAPIRequests  requestRecorder
	gatewayRequests    requestRecorder
	thirdPartyRequests requestRecorder
}

func newAPIHostTestHarness(t *testing.T, host, apiHost string) *apiHostTestHarness {
	t.Helper()

	harness := &apiHostTestHarness{}

	// First we stand up a TLS server that fakes the real GitHub API. It will record requests and respond with canned responses.
	fakeGitHub := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		harness.githubAPIRequests.record(req)
		w.Header().Set(contentType, jsonContentType)
		switch req.URL.Path {
		case "/http-client":
			_, _ = io.WriteString(w, `{"message":"http client response"}`)
		case "/direct-api-host":
			_, _ = io.WriteString(w, `{"message":"direct api host response"}`)
		case "/repos/cli/example-repository", "/api/v3/repos/cli/example-repository":
			_, _ = io.WriteString(w, `{"name":"example-repository"}`)
		case "/repositories", "/api/v3/repositories":
			if req.URL.Query().Get("page") == "2" {
				_, _ = io.WriteString(w, `[{"name":"example-repository-page-2"}]`)
				return
			}
			w.Header().Set("Link", fmt.Sprintf(`<https://%s%s?page=2>; rel="next"`, apiHost, req.URL.Path))
			_, _ = io.WriteString(w, `[{"name":"example-repository-page-1"}]`)
		case "/graphql", "/api/graphql":
			_, _ = io.WriteString(w, `{"data":{"viewer":{"login":"hubot"}}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(fakeGitHub.Close)

	target, err := url.Parse(fakeGitHub.URL)
	require.NoError(t, err)

	// Then we stand up a proxy server that will forward requests to the fake GitHub server, i.e. the api_host.
	// It will also record requests that it receives, so we can assert that it was called when we expected it to be.
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = fakeGitHub.Client().Transport

	gateway := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		harness.gatewayRequests.record(req)
		switch req.URL.Path {
		case "/redirect-to-third-party":
			http.Redirect(w, req, "https://"+thirdPartyHost+"/redirected", http.StatusFound)
			return
		case "/redirect-to-subdomain":
			http.Redirect(w, req, "https://subdomain.gw.example.net/redirected", http.StatusFound)
			return
		case "/redirect-to-canonical-host":
			http.Redirect(w, req, "https://"+host+"/redirected", http.StatusFound)
			return
		case "/redirect-to-canonical-subdomain":
			http.Redirect(w, req, "https://subdomain."+host+"/redirected", http.StatusFound)
			return
		}
		proxy.ServeHTTP(w, req)
	}))
	t.Cleanup(gateway.Close)

	// Finally we stand up a server representing an unrelated third party, which must never
	// receive the auth token, whether it is requested directly or arrived at via a redirect.
	thirdParty := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		harness.thirdPartyRequests.record(req)
		w.Header().Set(contentType, jsonContentType)
		_, _ = io.WriteString(w, `{"message":"third party response"}`)
	}))
	t.Cleanup(thirdParty.Close)

	// To allow us to use fake domain names that are representative, we use a custom transport
	// that rewrites the hostnames to point to our test servers.
	fakeAddress := fakeGitHub.Listener.Addr().String()
	gatewayAddress := gateway.Listener.Addr().String()
	dialMap := map[string]string{
		"api.github.com:443":           fakeAddress,
		"api.example.ghe.com:443":      fakeAddress,
		"ghes.example.com:443":         fakeAddress,
		"gw.example.net:443":           gatewayAddress,
		"subdomain.gw.example.net:443": gatewayAddress,
		host + ":443":                  fakeAddress,
		"subdomain." + host + ":443":   fakeAddress,
		thirdPartyHost + ":443":        thirdParty.Listener.Addr().String(),
	}
	harness.transport = &http.Transport{
		// We must turn off TLS verification because our test servers use self-signed certs.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			if mapped, ok := dialMap[address]; ok {
				address = mapped
			}
			return (&net.Dialer{}).DialContext(ctx, network, address)
		},
	}
	t.Cleanup(harness.transport.CloseIdleConnections)

	return harness
}

func newConfiguredAPIHostTest(t *testing.T, host, apiHost string) (*apiHostTestHarness, ClientOptions) {
	t.Helper()

	harness := newAPIHostTestHarness(t, host, apiHost)
	testutils.StubConfig(t, fmt.Sprintf("hosts:\n  %s:\n    api_host: %q\n", host, apiHost))
	return harness, ClientOptions{
		Host:      host,
		AuthToken: "test-token",
		Transport: harness.transport,
	}
}

func newCanonicalAPIHostTest(t *testing.T, apiHost string) (*apiHostTestHarness, ClientOptions) {
	t.Helper()

	harness := newAPIHostTestHarness(t, "github.com", apiHost)
	testutils.StubConfig(t, "")
	return harness, ClientOptions{
		Host:      "github.com",
		AuthToken: "test-token",
		Transport: harness.transport,
	}
}

func TestAPIHostRouting(t *testing.T) {
	const apiHost = "gw.example.net"

	tests := []struct {
		name        string
		host        string
		restPath    string
		pagePath    string
		graphqlPath string
	}{
		{
			name:        "github.com",
			host:        "github.com",
			restPath:    "/repos/cli/example-repository",
			pagePath:    "/repositories",
			graphqlPath: "/graphql",
		},
		{
			name:        "ghe.com tenancy",
			host:        "example.ghe.com",
			restPath:    "/repos/cli/example-repository",
			pagePath:    "/repositories",
			graphqlPath: "/graphql",
		},
		{
			name:        "GHES",
			host:        "ghes.example.com",
			restPath:    "/api/v3/repos/cli/example-repository",
			pagePath:    "/api/v3/repositories",
			graphqlPath: "/api/graphql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("HTTP client looks up api_host from config", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				httpClient, err := NewHTTPClient(opts)
				require.NoError(t, err)

				response, err := httpClient.Get("https://" + apiHost + "/http-client")
				require.NoError(t, err)
				require.NoError(t, response.Body.Close())

				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/http-client",
					host:          apiHost,
					authorization: "token test-token",
				})
			})

			t.Run("REST client looks up api_host from config", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				restClient, err := NewRESTClient(opts)
				require.NoError(t, err)

				var restResult struct {
					Name string `json:"name"`
				}
				require.NoError(t, restClient.Get("repos/cli/example-repository", &restResult))
				assert.Equal(t, "example-repository", restResult.Name)

				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodGet,
					path:          tt.restPath,
					host:          apiHost,
					authorization: "token test-token",
				})
			})

			t.Run("GraphQL client looks up api_host from config", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				graphQLClient, err := NewGraphQLClient(opts)
				require.NoError(t, err)

				var graphQLResult struct {
					Viewer struct {
						Login string `json:"login"`
					} `json:"viewer"`
				}
				require.NoError(t, graphQLClient.Do("query { viewer { login } }", nil, &graphQLResult))
				assert.Equal(t, "hubot", graphQLResult.Viewer.Login)

				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodPost,
					path:          tt.graphqlPath,
					host:          apiHost,
					authorization: "token test-token",
				})
			})

			t.Run("correctly provides token for direct api_host request", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				restClient, err := NewRESTClient(opts)
				require.NoError(t, err)

				var result struct {
					Message string `json:"message"`
				}
				require.NoError(t, restClient.Get("https://"+apiHost+"/direct-api-host", &result))
				assert.Equal(t, "direct api host response", result.Message)

				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/direct-api-host",
					host:          apiHost,
					authorization: "token test-token",
				})
			})

			t.Run("correctly provides tokens for pagination", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				restClient, err := NewRESTClient(opts)
				require.NoError(t, err)

				response, err := restClient.Request(http.MethodGet, "repositories?per_page=1", nil)
				require.NoError(t, err)
				require.NoError(t, response.Body.Close())

				nextPageURL := strings.TrimSuffix(strings.TrimPrefix(response.Header.Get("Link"), "<"), `>; rel="next"`)
				require.Equal(t, fmt.Sprintf("https://%s%s?page=2", apiHost, tt.pagePath), nextPageURL)

				var nextPageResult []struct {
					Name string `json:"name"`
				}
				require.NoError(t, restClient.Get(nextPageURL, &nextPageResult))
				require.Len(t, nextPageResult, 1)
				assert.Equal(t, "example-repository-page-2", nextPageResult[0].Name)

				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodGet,
					path:          tt.pagePath,
					rawQuery:      "page=2",
					host:          apiHost,
					authorization: "token test-token",
				})
			})

			t.Run("does not provide token to an unrelated host", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				restClient, err := NewRESTClient(opts)
				require.NoError(t, err)

				var result struct {
					Message string `json:"message"`
				}
				require.NoError(t, restClient.Get("https://"+thirdPartyHost+"/third-party", &result))
				assert.Equal(t, "third party response", result.Message)

				requireRequest(t, &harness.thirdPartyRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/third-party",
					host:          thirdPartyHost,
					authorization: "",
				})
			})

			t.Run("does not provide token when redirected off the api_host", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				httpClient, err := NewHTTPClient(opts)
				require.NoError(t, err)

				response, err := httpClient.Get("https://" + apiHost + "/redirect-to-third-party")
				require.NoError(t, err)
				require.NoError(t, response.Body.Close())

				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/redirect-to-third-party",
					host:          apiHost,
					authorization: "token test-token",
				})
				requireRequest(t, &harness.thirdPartyRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/redirected",
					host:          thirdPartyHost,
					authorization: "",
				})
			})

			t.Run("does not provide token when redirected to a subdomain of the api_host", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				httpClient, err := NewHTTPClient(opts)
				require.NoError(t, err)

				response, err := httpClient.Get("https://" + apiHost + "/redirect-to-subdomain")
				require.NoError(t, err)
				require.NoError(t, response.Body.Close())

				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/redirect-to-subdomain",
					host:          apiHost,
					authorization: "token test-token",
				})
				requireRequest(t, &harness.gatewayRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/redirected",
					host:          "subdomain." + apiHost,
					authorization: "",
				})
			})

			t.Run("provides token when redirected to the canonical host", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				httpClient, err := NewHTTPClient(opts)
				require.NoError(t, err)

				response, err := httpClient.Get("https://" + apiHost + "/redirect-to-canonical-host")
				require.NoError(t, err)
				require.NoError(t, response.Body.Close())

				requireRequest(t, &harness.githubAPIRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/redirected",
					host:          tt.host,
					authorization: "token test-token",
				})
			})

			t.Run("provides token when redirected to a canonical subdomain", func(t *testing.T) {
				harness, opts := newConfiguredAPIHostTest(t, tt.host, apiHost)
				httpClient, err := NewHTTPClient(opts)
				require.NoError(t, err)

				response, err := httpClient.Get("https://" + apiHost + "/redirect-to-canonical-subdomain")
				require.NoError(t, err)
				require.NoError(t, response.Body.Close())

				requireRequest(t, &harness.githubAPIRequests, recordedRequest{
					method:        http.MethodGet,
					path:          "/redirected",
					host:          "subdomain." + tt.host,
					authorization: "token test-token",
				})
			})
		})
	}

	t.Run("no override goes directly to canonical host", func(t *testing.T) {
		t.Run("HTTP client uses canonical host", func(t *testing.T) {
			harness, opts := newCanonicalAPIHostTest(t, apiHost)
			httpClient, err := NewHTTPClient(opts)
			require.NoError(t, err)

			response, err := httpClient.Get("https://api.github.com/http-client")
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())

			requireRequest(t, &harness.githubAPIRequests, recordedRequest{
				method:        http.MethodGet,
				path:          "/http-client",
				host:          "api.github.com",
				authorization: "token test-token",
			})
			assert.Empty(t, harness.gatewayRequests.recordedRequests())
		})

		t.Run("REST client uses canonical host", func(t *testing.T) {
			harness, opts := newCanonicalAPIHostTest(t, apiHost)
			restClient, err := NewRESTClient(opts)
			require.NoError(t, err)

			var restResult struct {
				Name string `json:"name"`
			}
			require.NoError(t, restClient.Get("repos/cli/example-repository", &restResult))
			assert.Equal(t, "example-repository", restResult.Name)

			requireRequest(t, &harness.githubAPIRequests, recordedRequest{
				method:        http.MethodGet,
				path:          "/repos/cli/example-repository",
				host:          "api.github.com",
				authorization: "token test-token",
			})
			assert.Empty(t, harness.gatewayRequests.recordedRequests())
		})

		t.Run("GraphQL client uses canonical host", func(t *testing.T) {
			harness, opts := newCanonicalAPIHostTest(t, apiHost)
			graphQLClient, err := NewGraphQLClient(opts)
			require.NoError(t, err)

			var graphQLResult struct {
				Viewer struct {
					Login string `json:"login"`
				} `json:"viewer"`
			}
			require.NoError(t, graphQLClient.Do("query { viewer { login } }", nil, &graphQLResult))
			assert.Equal(t, "hubot", graphQLResult.Viewer.Login)

			requireRequest(t, &harness.githubAPIRequests, recordedRequest{
				method:        http.MethodPost,
				path:          "/graphql",
				host:          "api.github.com",
				authorization: "token test-token",
			})
			assert.Empty(t, harness.gatewayRequests.recordedRequests())
		})
	})
}
