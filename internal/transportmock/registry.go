package transportmock

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
)

type Registry struct {
	mu       sync.Mutex
	Stubs    []*stub
	Requests []*http.Request
}

func NewRegistry(t *testing.T) *Registry {
	reg := Registry{}
	t.Cleanup(func() { reg.Verify(t) })
	return &reg
}

func (r *Registry) Register(name string, matcher Matcher, responder Responder) {
	r.Stubs = append(r.Stubs, &stub{
		name:      name,
		matcher:   matcher,
		responder: responder,
	})
}

func (r *Registry) Verify(t *testing.T) {
	unmatched := []string{}
	for _, s := range r.Stubs {
		if !s.matched {
			unmatched = append(unmatched, s.name)
		}
	}
	if len(unmatched) > 0 {
		t.Helper()
		t.Errorf("%d unmatched stubs: %s", len(unmatched), unmatched)
	}
}

// Registry satisfies http.RoundTripper interface.
func (r *Registry) RoundTrip(req *http.Request) (*http.Response, error) {
	var stub *stub
	r.mu.Lock()
	for _, s := range r.Stubs {
		if s.matched || !s.matcher(req) {
			continue
		}
		if stub != nil {
			r.mu.Unlock()
			return nil, fmt.Errorf("both %s stub and %s stub matched %v", s, stub, req)
		}
		stub = s
	}
	if stub != nil {
		stub.matched = true
	}
	if stub == nil {
		r.mu.Unlock()
		return nil, fmt.Errorf("no registered stubs matched %v", req)
	}
	r.Requests = append(r.Requests, req)
	r.mu.Unlock()
	return stub.responder(req)
}
