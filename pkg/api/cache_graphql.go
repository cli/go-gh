package api

import (
	"encoding/json"
	"io"
	"net/http"
)

// isCacheableGraphQLRequest decides whether a POST /graphql request is safe
// to persist in the cache. It is fail-closed: any parse difficulty, missing
// query field, or evidence of a mutation or subscription operation returns
// false (do not cache).
//
// This peeks at req.Body via copyStream so the request body can still be
// consumed by downstream readers (cacheKey + the underlying RoundTripper).
func isCacheableGraphQLRequest(req *http.Request) bool {
	if req.Body == nil {
		return false
	}

	var bodyCopy io.ReadCloser
	req.Body, bodyCopy = copyStream(req.Body)
	defer bodyCopy.Close()

	bodyBytes, err := io.ReadAll(bodyCopy)
	if err != nil {
		return false
	}

	var doc struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(bodyBytes, &doc); err != nil {
		return false
	}
	if doc.Query == "" {
		return false
	}

	return !graphqlContainsMutationOrSubscription(doc.Query)
}

// graphqlContainsMutationOrSubscription scans a GraphQL document and returns
// true if any top-level operation definition is a mutation or a subscription,
// or if the document is shaped in a way the scanner cannot prove is a query.
//
// The scanner is intentionally tiny: it tokenizes just enough to track brace,
// paren, and bracket depth, skip strings and comments, and recognize
// operation-type keywords at definition boundaries. False positives (refusing
// to cache a valid query, e.g. a query whose first directive happens to be
// named @mutation) are safe in this direction; they just degrade to "no
// cache" for that one request. False negatives would cache mutation results,
// which is a correctness bug, so anything ambiguous returns true.
func graphqlContainsMutationOrSubscription(query string) bool {
	pos := 0
	n := len(query)

	if n >= 3 && query[0] == 0xEF && query[1] == 0xBB && query[2] == 0xBF {
		pos = 3
	}

	braceDepth := 0
	parenDepth := 0
	bracketDepth := 0
	expectDefinition := true

	atTopLevel := func() bool {
		return braceDepth == 0 && parenDepth == 0 && bracketDepth == 0
	}

	for pos < n {
		c := query[pos]

		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' {
			pos++
			continue
		}

		if c == '#' {
			for pos < n && query[pos] != '\n' {
				pos++
			}
			continue
		}

		if c == '"' {
			if pos+2 < n && query[pos+1] == '"' && query[pos+2] == '"' {
				pos += 3
				closed := false
				for pos+2 < n {
					if query[pos] == '"' && query[pos+1] == '"' && query[pos+2] == '"' {
						pos += 3
						closed = true
						break
					}
					pos++
				}
				if !closed {
					return true
				}
			} else {
				pos++
				closed := false
				for pos < n && query[pos] != '"' {
					if query[pos] == '\\' && pos+1 < n {
						pos += 2
						continue
					}
					if query[pos] == '\n' {
						return true
					}
					pos++
				}
				if pos < n && query[pos] == '"' {
					pos++
					closed = true
				}
				if !closed {
					return true
				}
			}
			expectDefinition = false
			continue
		}

		switch c {
		case '{':
			braceDepth++
			expectDefinition = false
			pos++
			continue
		case '}':
			if braceDepth > 0 {
				braceDepth--
			}
			if atTopLevel() {
				expectDefinition = true
			}
			pos++
			continue
		case '(':
			parenDepth++
			pos++
			continue
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
			pos++
			continue
		case '[':
			bracketDepth++
			pos++
			continue
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			pos++
			continue
		}

		if isGraphQLNameStart(c) {
			start := pos
			for pos < n && isGraphQLNameCont(query[pos]) {
				pos++
			}
			name := query[start:pos]

			if expectDefinition && atTopLevel() {
				switch name {
				case "mutation", "subscription":
					return true
				case "query", "fragment":
					expectDefinition = false
					continue
				default:
					return true
				}
			}
			expectDefinition = false
			continue
		}

		pos++
	}

	return false
}

func isGraphQLNameStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isGraphQLNameCont(c byte) bool {
	return isGraphQLNameStart(c) || (c >= '0' && c <= '9')
}
