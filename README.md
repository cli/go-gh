# go-gh

A Go module for CLI Go applications and [gh extensions]() that want a convenient way to interact with [gh](), and the GitHub API using `gh` environment configuration.

## Installation
```bash
go get https://github.com/cli/go-gh
```

## Usage
```golang
import (
	"fmt"
	"github.com/cli/go-gh"
)

// Execute `gh issue list -R cli/cli`, and print the output.
func main() {
	args := []string{"issue", "list", "-R", "cli/cli"}
	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(stdOut.String())
	fmt.Println(stdErr.String())
}
```

See [examples folder]() for more use cases.
