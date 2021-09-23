# go-gh

A Go module for CLI Go applications and [gh extensions][extensions] that want a convenient way to interact with [gh][], and the GitHub API using [gh][] environment configuration.

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

See [examples][examples] for more use cases.

## Contributing

If anything feels off, or if you feel that some functionality is missing, please check out the [contributing page][contributing]. There you will find instructions for sharing your feedback, and submitting pull requests to the project.

[extensions]: https://github.com/topics/gh-extension
[gh]: https://github.com/cli/cli
[examples]: ./example_gh_test.go
[contributing]: ./.github/CONTRIBUTING.md
