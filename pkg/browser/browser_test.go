package browser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cli/go-gh/v2/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GH_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "%v", os.Args[3:])
	os.Exit(0)
}

// TestBrowse ensures supported URLs are opened by the browser launcher.
// Running package tests in VSCode will cause this to fail due to use of
// `-coverageprofile` flag without `GOCOVERDIR` env var.
func TestBrowse(t *testing.T) {
	type browseTest struct {
		name     string
		url      string
		launcher string
		expected string
		setup    func(*testing.T) error
		wantErr  bool
	}

	tests := []browseTest{
		{
			name:     "Explicit `http` URL works",
			url:      "http://github.com",
			launcher: fmt.Sprintf("%q -test.run=TestHelperProcess -- explicit http", os.Args[0]),
			expected: "[explicit http http://github.com]",
		},
		{
			name:     "Explicit `https` URL works",
			url:      "https://github.com",
			launcher: fmt.Sprintf("%q -test.run=TestHelperProcess -- explicit https", os.Args[0]),
			expected: "[explicit https https://github.com]",
		},
		{
			name:     "Explicit `HTTPS` URL works",
			url:      "HTTPS://github.com",
			launcher: fmt.Sprintf("%q -test.run=TestHelperProcess -- explicit HTTPS", os.Args[0]),
			expected: "[explicit HTTPS https://github.com]",
		},
		{
			name:     "Explicit `vscode` URL works",
			url:      "vscode:extension/GitHub.copilot",
			launcher: fmt.Sprintf("%q -test.run=TestHelperProcess -- explicit vscode", os.Args[0]),
			expected: "[explicit vscode vscode:extension/GitHub.copilot]",
		},
		{
			name:     "Explicit `vscode-insiders` URL works",
			url:      "vscode-insiders:extension/GitHub.copilot",
			launcher: fmt.Sprintf("%q -test.run=TestHelperProcess -- explicit vscode-insiders", os.Args[0]),
			expected: "[explicit vscode-insiders vscode-insiders:extension/GitHub.copilot]",
		},
		{
			name:     "Implicit `https` URL works",
			url:      "github.com",
			launcher: fmt.Sprintf("%q -test.run=TestHelperProcess -- implicit https", os.Args[0]),
			expected: "[implicit https https://github.com]",
		},
		{
			name:    "Explicit absolute `file://` URL errors",
			url:     "file:///System/Applications/Calculator.app",
			wantErr: true,
		},
	}

	// Setup additional test scenarios covering OS-specific executables and directories
	// that should be installed on maintainer workstations and GitHub hosted runners.
	switch runtime.GOOS {
	case "windows":
		tests = append(tests, []browseTest{
			{
				name:    "Explicit absolute Windows file URL errors",
				url:     `C:\Windows\System32\cmd.exe`,
				wantErr: true,
			},
			{
				name:    "Explicit absolute Windows directory URL errors",
				url:     `C:\Windows\System32`,
				wantErr: true,
			},
		}...)
	// Default should handle common Unix/Linux scenarios including Mac OS.
	default:
		tests = append(tests, []browseTest{
			{
				name:    "Implicit absolute Unix/Linux file URL errors",
				url:     "/bin/bash",
				wantErr: true,
			},
			{
				name:    "Implicit absolute Unix/Linux directory URL errors",
				url:     "/bin",
				wantErr: true,
			},
			{
				name: "Implicit relative Unix/Linux file URL errors",
				url:  "poc.command",
				setup: func(t *testing.T) error {
					// Setup a temporary directory to stage content and execute the test within,
					// ensure the test's original working directory is restored after.
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}

					tempDir := t.TempDir()
					err = os.Chdir(tempDir)
					if err != nil {
						return err
					}

					t.Cleanup(func() {
						_ = os.Chdir(cwd)
					})

					// Create content for local file URL testing
					path := filepath.Join(tempDir, "poc.command")
					err = os.WriteFile(path, []byte("#!/bin/bash\necho hello"), 0755)
					if err != nil {
						return err
					}

					return nil
				},
				wantErr: true,
			},
			{
				name: "Implicit relative Unix/Linux directory URL errors",
				url:  "Fake.app",
				setup: func(t *testing.T) error {
					// Setup a temporary directory to stage content and execute the test within,
					// ensure the test's original working directory is restored after.
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}

					tempDir := t.TempDir()
					err = os.Chdir(tempDir)
					if err != nil {
						return err
					}

					t.Cleanup(func() {
						_ = os.Chdir(cwd)
					})

					// Create content for local directory URL testing
					path := filepath.Join(tempDir, "Fake.app")
					err = os.Mkdir(path, 0755)
					if err != nil {
						return err
					}

					path = filepath.Join(path, "poc.command")
					err = os.WriteFile(path, []byte("#!/bin/bash\necho hello"), 0755)
					if err != nil {
						return err
					}

					return nil
				},
				wantErr: true,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup(t)
				require.NoError(t, err)
			}

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			b := Browser{launcher: tt.launcher, stdout: stdout, stderr: stderr}
			err := b.browse(tt.url, []string{"GH_WANT_HELPER_PROCESS=1"})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, stdout.String())
				assert.Equal(t, "", stderr.String())
			}
		})
	}
}

func TestResolveLauncher(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		config       *config.Config
		wantLauncher string
	}{
		{
			name: "GH_BROWSER set",
			env: map[string]string{
				"GH_BROWSER": "GH_BROWSER",
			},
			wantLauncher: "GH_BROWSER",
		},
		{
			name:         "config browser set",
			config:       config.ReadFromString("browser: CONFIG_BROWSER"),
			wantLauncher: "CONFIG_BROWSER",
		},
		{
			name: "BROWSER set",
			env: map[string]string{
				"BROWSER": "BROWSER",
			},
			wantLauncher: "BROWSER",
		},
		{
			name: "GH_BROWSER and config browser set",
			env: map[string]string{
				"GH_BROWSER": "GH_BROWSER",
			},
			config:       config.ReadFromString("browser: CONFIG_BROWSER"),
			wantLauncher: "GH_BROWSER",
		},
		{
			name: "config browser and BROWSER set",
			env: map[string]string{
				"BROWSER": "BROWSER",
			},
			config:       config.ReadFromString("browser: CONFIG_BROWSER"),
			wantLauncher: "CONFIG_BROWSER",
		},
		{
			name: "GH_BROWSER and BROWSER set",
			env: map[string]string{
				"BROWSER":    "BROWSER",
				"GH_BROWSER": "GH_BROWSER",
			},
			wantLauncher: "GH_BROWSER",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != nil {
				for k, v := range tt.env {
					t.Setenv(k, v)
				}
			}
			if tt.config != nil {
				old := config.Read
				config.Read = func(_ *config.Config) (*config.Config, error) {
					return tt.config, nil
				}
				defer func() { config.Read = old }()
			}
			launcher := resolveLauncher()
			assert.Equal(t, tt.wantLauncher, launcher)
		})
	}
}
