# Go library for the GitHub CLI

`go-gh` is a collection of Go modules to make authoring [GitHub CLI extensions][extensions] easier.

Modules from this library will obey GitHub CLI conventions by default:

- [`repository.Current()`](https://pkg.go.dev/github.com/cli/go-gh/v2/pkg/repository#current) respects the value of the `GH_REPO` environment variable and reads from git remote configuration as fallback.

- GitHub API requests will be authenticated using the same mechanism as `gh`, i.e. using the values of `GH_TOKEN` and `GH_HOST` environment variables and falling back to the user's stored OAuth token.

- [Terminal capabilities](https://pkg.go.dev/github.com/cli/go-gh/v2/pkg/term) are determined by taking environment variables `GH_FORCE_TTY`, `NO_COLOR`, `CLICOLOR`, etc. into account.

- Generating [table](https://pkg.go.dev/github.com/cli/go-gh/v2/pkg/tableprinter) or [Go template](https://pkg.go.dev/github.com/cli/go-gh/pkg/template) output uses the same engine as gh.

- The [`browser`](https://pkg.go.dev/github.com/cli/go-gh/v2/pkg/browser) module activates the user's preferred web browser.

## Usage

See the full `go-gh`  [reference documentation](https://pkg.go.dev/github.com/cli/go-gh/v2) for more information

```golang
package main

import (
	"fmt"
	"log"
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
)

func main() {
	// These examples assume `gh` is installed and has been authenticated.

	// Shell out to a gh command and read its output.
	issueList, _, err := gh.Exec("issue", "list", "--repo", "cli/cli", "--limit", "5")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(issueList.String())

	// Use an API client to retrieve repository tags.
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatal(err)
	}
	response := []struct{
		Name string
	}{}
	err = client.Get("repos/cli/cli/tags", &response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}
```

See [examples][] for more demonstrations of usage.

## Contributing

If anything feels off, or if you feel that some functionality is missing, please check out our [contributing docs][contributing]. There you will find instructions for sharing your feedback and for submitting pull requests to the project. Thank you!

[extensions]: https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions
[examples]: ./example_gh_test.go
[contributing]: ./.github/CONTRIBUTING.md
