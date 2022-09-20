package git

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "scp-like",
			url:  "git@example.com:owner/repo",
			want: true,
		},
		{
			name: "scp-like with no user",
			url:  "example.com:owner/repo",
			want: false,
		},
		{
			name: "ssh",
			url:  "ssh://git@example.com/owner/repo",
			want: true,
		},
		{
			name: "git",
			url:  "git://example.com/owner/repo",
			want: true,
		},
		{
			name: "git with extension",
			url:  "git://example.com/owner/repo.git",
			want: true,
		},
		{
			name: "git+ssh",
			url:  "git+ssh://git@example.com/owner/repo.git",
			want: true,
		},
		{
			name: "git+https",
			url:  "git+https://example.com/owner/repo.git",
			want: true,
		},
		{
			name: "http",
			url:  "http://example.com/owner/repo.git",
			want: true,
		},
		{
			name: "https",
			url:  "https://example.com/owner/repo.git",
			want: true,
		},
		{
			name: "no protocol",
			url:  "example.com/owner/repo",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsURL(tt.url))
		})
	}
}

func TestParseURL(t *testing.T) {
	type url struct {
		Scheme string
		User   string
		Host   string
		Path   string
	}

	tests := []struct {
		name    string
		url     string
		want    url
		wantErr bool
	}{
		{
			name: "HTTPS",
			url:  "https://example.com/owner/repo.git",
			want: url{
				Scheme: "https",
				User:   "",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "HTTP",
			url:  "http://example.com/owner/repo.git",
			want: url{
				Scheme: "http",
				User:   "",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "git",
			url:  "git://example.com/owner/repo.git",
			want: url{
				Scheme: "git",
				User:   "",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "ssh",
			url:  "ssh://git@example.com/owner/repo.git",
			want: url{
				Scheme: "ssh",
				User:   "git",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "ssh with port",
			url:  "ssh://git@example.com:443/owner/repo.git",
			want: url{
				Scheme: "ssh",
				User:   "git",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "git+ssh",
			url:  "git+ssh://example.com/owner/repo.git",
			want: url{
				Scheme: "ssh",
				User:   "",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "git+https",
			url:  "git+https://example.com/owner/repo.git",
			want: url{
				Scheme: "https",
				User:   "",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "scp-like",
			url:  "git@example.com:owner/repo.git",
			want: url{
				Scheme: "ssh",
				User:   "git",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "scp-like, leading slash",
			url:  "git@example.com:/owner/repo.git",
			want: url{
				Scheme: "ssh",
				User:   "git",
				Host:   "example.com",
				Path:   "/owner/repo.git",
			},
		},
		{
			name: "file protocol",
			url:  "file:///example.com/owner/repo.git",
			want: url{
				Scheme: "file",
				User:   "",
				Host:   "",
				Path:   "/example.com/owner/repo.git",
			},
		},
		{
			name: "file path",
			url:  "/example.com/owner/repo.git",
			want: url{
				Scheme: "",
				User:   "",
				Host:   "",
				Path:   "/example.com/owner/repo.git",
			},
		},
		{
			name: "Windows file path",
			url:  "C:\\example.com\\owner\\repo.git",
			want: url{
				Scheme: "c",
				User:   "",
				Host:   "",
				Path:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := ParseURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Scheme, u.Scheme)
			assert.Equal(t, tt.want.User, u.User.Username())
			assert.Equal(t, tt.want.Host, u.Host)
			assert.Equal(t, tt.want.Path, u.Path)
		})
	}
}

func TestRepoInfoFromURL(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantHost   string
		wantOwner  string
		wantRepo   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "github.com URL",
			input:     "https://github.com/monalisa/octo-cat.git",
			wantHost:  "github.com",
			wantOwner: "monalisa",
			wantRepo:  "octo-cat",
		},
		{
			name:      "github.com URL with trailing slash",
			input:     "https://github.com/monalisa/octo-cat/",
			wantHost:  "github.com",
			wantOwner: "monalisa",
			wantRepo:  "octo-cat",
		},
		{
			name:      "www.github.com URL",
			input:     "http://www.GITHUB.com/monalisa/octo-cat.git",
			wantHost:  "github.com",
			wantOwner: "monalisa",
			wantRepo:  "octo-cat",
		},
		{
			name:       "too many path components",
			input:      "https://github.com/monalisa/octo-cat/pulls",
			wantErr:    true,
			wantErrMsg: "invalid path: /monalisa/octo-cat/pulls",
		},
		{
			name:      "non-GitHub hostname",
			input:     "https://example.com/one/two",
			wantHost:  "example.com",
			wantOwner: "one",
			wantRepo:  "two",
		},
		{
			name:       "filesystem path",
			input:      "/path/to/file",
			wantErr:    true,
			wantErrMsg: "no hostname detected",
		},
		{
			name:       "filesystem path with scheme",
			input:      "file:///path/to/file",
			wantErr:    true,
			wantErrMsg: "no hostname detected",
		},
		{
			name:      "github.com SSH URL",
			input:     "ssh://github.com/monalisa/octo-cat.git",
			wantHost:  "github.com",
			wantOwner: "monalisa",
			wantRepo:  "octo-cat",
		},
		{
			name:      "github.com HTTPS+SSH URL",
			input:     "https+ssh://github.com/monalisa/octo-cat.git",
			wantHost:  "github.com",
			wantOwner: "monalisa",
			wantRepo:  "octo-cat",
		},
		{
			name:      "github.com git URL",
			input:     "git://github.com/monalisa/octo-cat.git",
			wantHost:  "github.com",
			wantOwner: "monalisa",
			wantRepo:  "octo-cat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			assert.NoError(t, err)
			host, owner, repo, err := RepoInfoFromURL(u)
			if tt.wantErr {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantOwner, owner)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}
