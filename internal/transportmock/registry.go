package transportmock

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
	"testing"
)

type Registry struct {
	mu       sync.Mutex
	stubs    []*stub
	requests []*http.Request
}

func NewRegistry(t *testing.T) *Registry {
	reg := Registry{}
	t.Cleanup(func() { reg.Verify(t) })
	return &reg
}

func (r *Registry) Register(name string, matcher Matcher, responder Responder) {
	r.stubs = append(r.stubs, &stub{
		name:      name,
		matcher:   matcher,
		responder: responder,
	})
}

func (r *Registry) Requests() []*http.Request {
	return r.requests
}

func (r *Registry) Verify(t *testing.T) {
	unmatched := []string{}
	for _, s := range r.stubs {
		if !s.matched {
			unmatched = append(unmatched, s.name)
		}
	}
	if len(unmatched) > 0 {
		t.Helper()
		t.Errorf("%d unmatched stubs: %s", len(unmatched), unmatched)
	}
}

func (r *Registry) RoundTrip(req *http.Request) (*http.Response, error) {
	var stub *stub
	r.mu.Lock()
	for _, s := range r.stubs {
		if s.matched || !s.matcher(req) {
			continue
		}
		if stub != nil {
			r.mu.Unlock()
			dr, _ := httputil.DumpRequestOut(req, true)
			return nil, fmt.Errorf("both %s stub and %s stub matched request\n%s", stub, s, dr)
		}
		stub = s
	}
	if stub != nil {
		stub.matched = true
	}
	if stub == nil {
		r.mu.Unlock()
		dr, _ := httputil.DumpRequestOut(req, true)
		return nil, fmt.Errorf("no stubs matched request\n%s", dr)
	}
	r.requests = append(r.requests, req)
	r.mu.Unlock()
	return stub.responder(req)
}
