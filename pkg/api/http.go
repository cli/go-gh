package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/henvic/httpretty"
)

const (
	accept          = "Accept"
	authorization   = "Authorization"
	contentType     = "Content-Type"
	defaultHostname = "github.com"
	jsonContentType = "application/json; charset=utf-8"
	timeZone        = "Time-Zone"
	userAgent       = "User-Agent"
)

var jsonTypeRE = regexp.MustCompile(`[/+]json($|;)`)
var timeZoneNames = map[int]string{
	-39600: "Pacific/Niue",
	-36000: "Pacific/Honolulu",
	-34200: "Pacific/Marquesas",
	-32400: "America/Anchorage",
	-28800: "America/Los_Angeles",
	-25200: "America/Chihuahua",
	-21600: "America/Chicago",
	-18000: "America/Bogota",
	-14400: "America/Caracas",
	-12600: "America/St_Johns",
	-10800: "America/Argentina/Buenos_Aires",
	-7200:  "Atlantic/South_Georgia",
	-3600:  "Atlantic/Cape_Verde",
	0:      "Europe/London",
	3600:   "Europe/Amsterdam",
	7200:   "Europe/Athens",
	10800:  "Europe/Istanbul",
	12600:  "Asia/Tehran",
	14400:  "Asia/Dubai",
	16200:  "Asia/Kabul",
	18000:  "Asia/Tashkent",
	19800:  "Asia/Kolkata",
	20700:  "Asia/Kathmandu",
	21600:  "Asia/Dhaka",
	23400:  "Asia/Rangoon",
	25200:  "Asia/Bangkok",
	28800:  "Asia/Manila",
	31500:  "Australia/Eucla",
	32400:  "Asia/Tokyo",
	34200:  "Australia/Darwin",
	36000:  "Australia/Brisbane",
	37800:  "Australia/Adelaide",
	39600:  "Pacific/Guadalcanal",
	43200:  "Pacific/Nauru",
	46800:  "Pacific/Auckland",
	49500:  "Pacific/Chatham",
	50400:  "Pacific/Kiritimati",
}

type ClientOptions struct {
	AuthToken   string
	CacheDir    string
	CacheTTL    time.Duration
	EnableCache bool
	Headers     map[string]string
	Log         io.Writer
	Timeout     time.Duration
	Transport   http.RoundTripper
}

func newHTTPClient(opts ClientOptions) http.Client {
	transport := http.DefaultTransport
	if opts.Transport != nil {
		transport = opts.Transport
	}

	// header RoundTripper
	transport = newHeaderRoundTripper(opts.AuthToken, opts.Headers, transport)

	// log RoundTripper
	if opts.Log != nil {
		logger := &httpretty.Logger{
			Time:            true,
			TLS:             false,
			Colors:          false,
			RequestHeader:   true,
			RequestBody:     true,
			ResponseHeader:  true,
			ResponseBody:    true,
			Formatters:      []httpretty.Formatter{&httpretty.JSONFormatter{}},
			MaxResponseBody: 10000,
		}
		logger.SetOutput(opts.Log)
		logger.SetBodyFilter(func(h http.Header) (skip bool, err error) {
			return !inspectableMIMEType(h.Get(contentType)), nil
		})
		transport = logger.RoundTripper(transport)
	}

	// cache RoundTripper
	if opts.EnableCache {
		if opts.CacheDir == "" {
			opts.CacheDir = filepath.Join(os.TempDir(), "gh-cli-cache")
		}
		if opts.CacheTTL == 0 {
			opts.CacheTTL = time.Hour * 24
		}
		c := cache{dir: opts.CacheDir, ttl: opts.CacheTTL}
		transport = c.RoundTripper(transport)
	}

	return http.Client{Transport: transport, Timeout: opts.Timeout}
}

type httpError struct {
	StatusCode  int
	RequestURL  *url.URL
	Message     string
	OAuthScopes string
	Errors      []httpErrorItem
}

type httpErrorItem struct {
	Message  string
	Resource string
	Field    string
	Code     string
}

func (err httpError) Error() string {
	if msgs := strings.SplitN(err.Message, "\n", 2); len(msgs) > 1 {
		return fmt.Sprintf("HTTP %d: %s (%s)\n%s", err.StatusCode, msgs[0], err.RequestURL, msgs[1])
	} else if err.Message != "" {
		return fmt.Sprintf("HTTP %d: %s (%s)", err.StatusCode, err.Message, err.RequestURL)
	}
	return fmt.Sprintf("HTTP %d (%s)", err.StatusCode, err.RequestURL)
}

func handleHTTPError(resp *http.Response) error {
	httpError := httpError{
		StatusCode:  resp.StatusCode,
		RequestURL:  resp.Request.URL,
		OAuthScopes: resp.Header.Get("X-Oauth-Scopes"),
	}

	if !jsonTypeRE.MatchString(resp.Header.Get(contentType)) {
		httpError.Message = resp.Status
		return httpError
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httpError.Message = err.Error()
		return httpError
	}

	var parsedBody struct {
		Message string `json:"message"`
		Errors  []json.RawMessage
	}
	if err := json.Unmarshal(body, &parsedBody); err != nil {
		return httpError
	}

	var messages []string
	if parsedBody.Message != "" {
		messages = append(messages, parsedBody.Message)
	}
	for _, raw := range parsedBody.Errors {
		switch raw[0] {
		case '"':
			var errString string
			_ = json.Unmarshal(raw, &errString)
			messages = append(messages, errString)
			httpError.Errors = append(httpError.Errors, httpErrorItem{Message: errString})
		case '{':
			var errInfo httpErrorItem
			_ = json.Unmarshal(raw, &errInfo)
			msg := errInfo.Message
			if errInfo.Code != "" && errInfo.Code != "custom" {
				msg = fmt.Sprintf("%s.%s %s", errInfo.Resource, errInfo.Field, errorCodeToMessage(errInfo.Code))
			}
			if msg != "" {
				messages = append(messages, msg)
			}
			httpError.Errors = append(httpError.Errors, errInfo)
		}
	}
	httpError.Message = strings.Join(messages, "\n")

	return httpError
}

// https://docs.github.com/en/rest/overview/resources-in-the-rest-api#client-errors
func errorCodeToMessage(code string) string {
	switch code {
	case "missing", "missing_field":
		return "is missing"
	case "invalid", "unprocessable":
		return "is invalid"
	case "already_exists":
		return "already exists"
	default:
		return code
	}
}

func inspectableMIMEType(t string) bool {
	return strings.HasPrefix(t, "text/") || jsonTypeRE.MatchString(t)
}

func isEnterprise(host string) bool {
	return host != defaultHostname
}

type headerRoundTripper struct {
	headers map[string]string
	rt      http.RoundTripper
}

func newHeaderRoundTripper(authToken string, headers map[string]string, rt http.RoundTripper) http.RoundTripper {
	if headers == nil {
		headers = map[string]string{}
	}
	if headers[contentType] == "" {
		headers[contentType] = jsonContentType
	}
	if headers[userAgent] == "" {
		headers[userAgent] = "go-gh" //TODO: Add some default user agent maybe using debug.ReadBuildInfo()
	}
	if headers[authorization] == "" && authToken != "" {
		headers[authorization] = fmt.Sprintf("token %s", authToken)
	}
	if headers[timeZone] == "" {
		headers[timeZone] = currentTimeZone()
	}
	if headers[accept] == "" {
		a := "application/vnd.github.merge-info-preview+json" // PullRequest.mergeStateStatus
		a += ", application/vnd.github.nebula-preview"        // visibility when RESTing repos into an org
		a += ", application/vnd.github.antiope-preview"       // Commit.statusCheckRollup for old GHES versions
		a += ", application/vnd.github.shadow-cat-preview"    // PullRequest.isDraft for old GHES versions
		headers[accept] = a
	}
	return headerRoundTripper{headers: headers, rt: rt}
}

func (hrt headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range hrt.headers {
		req.Header.Set(k, v)
	}
	return hrt.rt.RoundTrip(req)
}

func currentTimeZone() string {
	tz := time.Local.String()
	if tz == "Local" {
		_, offset := time.Now().Zone()
		tz = timeZoneNames[offset]
	}
	return tz
}
