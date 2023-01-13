package gh_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	graphql "github.com/cli/shurcooL-graphql"
)

// Execute 'gh issue list -R cli/cli', and print the output.
func ExampleExec() {
	args := []string{"issue", "list", "-R", "cli/cli"}
	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stdOut.String())
	fmt.Println(stdErr.String())
}

// Get tags from cli/cli repository using REST API.
func ExampleDefaultRESTClient() {
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatal(err)
	}
	response := []struct{ Name string }{}
	err = client.Get("repos/cli/cli/tags", &response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}

// Get tags from cli/cli repository using REST API.
// Specifying host, auth token, headers and logging to stdout.
func ExampleRESTClient() {
	opts := api.ClientOptions{
		Host:      "github.com",
		AuthToken: "xxxxxxxxxx", // Replace with valid auth token.
		Headers:   map[string]string{"Time-Zone": "America/Los_Angeles"},
		Log:       os.Stdout,
	}
	client, err := api.NewRESTClient(opts)
	if err != nil {
		log.Fatal(err)
	}
	response := []struct{ Name string }{}
	err = client.Get("repos/cli/cli/tags", &response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}

// Get release asset from cli/cli repository using REST API.
func ExampleRESTClient_request() {
	opts := api.ClientOptions{
		Headers: map[string]string{"Accept": "application/octet-stream"},
	}
	client, err := api.NewRESTClient(opts)
	if err != nil {
		log.Fatal(err)
	}
	// URL to cli/cli release v2.14.2 checksums.txt
	assetURL := "repos/cli/cli/releases/assets/71589494"
	response, err := client.Request(http.MethodGet, assetURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	f, err := os.CreateTemp("", "*_checksums.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = io.Copy(f, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Asset downloaded to %s\n", f.Name())
}

// Get releases from cli/cli repository using REST API with paginated results.
func ExampleRESTClient_pagination() {
	var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)
	findNextPage := func(response *http.Response) (string, bool) {
		for _, m := range linkRE.FindAllStringSubmatch(response.Header.Get("Link"), -1) {
			if len(m) > 2 && m[2] == "next" {
				return m[1], true
			}
		}
		return "", false
	}
	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatal(err)
	}
	requestPath := "repos/cli/cli/releases"
	page := 1
	for {
		response, err := client.Request(http.MethodGet, requestPath, nil)
		if err != nil {
			log.Fatal(err)
		}
		data := []struct{ Name string }{}
		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&data)
		if err != nil {
			log.Fatal(err)
		}
		if err := response.Body.Close(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Page: %d\n", page)
		fmt.Println(data)
		var hasNextPage bool
		if requestPath, hasNextPage = findNextPage(response); !hasNextPage {
			break
		}
		page++
	}
}

// Query tags from cli/cli repository using GQL API.
func ExampleDefaultGQLClient() {
	client, err := api.DefaultGQLClient()
	if err != nil {
		log.Fatal(err)
	}
	var query struct {
		Repository struct {
			Refs struct {
				Nodes []struct {
					Name string
				}
			} `graphql:"refs(refPrefix: $refPrefix, last: $last)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"refPrefix": graphql.String("refs/tags/"),
		"last":      graphql.Int(30),
		"owner":     graphql.String("cli"),
		"name":      graphql.String("cli"),
	}
	err = client.Query("RepositoryTags", &query, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(query)
}

// Query tags from cli/cli repository using GQL API.
// Enable caching and request timeout.
func ExampleGQLClient() {
	opts := api.ClientOptions{
		EnableCache: true,
		Timeout:     5 * time.Second,
	}
	client, err := api.NewGQLClient(opts)
	if err != nil {
		log.Fatal(err)
	}
	var query struct {
		Repository struct {
			Refs struct {
				Nodes []struct {
					Name string
				}
			} `graphql:"refs(refPrefix: $refPrefix, last: $last)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"refPrefix": graphql.String("refs/tags/"),
		"last":      graphql.Int(30),
		"owner":     graphql.String("cli"),
		"name":      graphql.String("cli"),
	}
	err = client.Query("RepositoryTags", &query, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(query)
}

// Add a star to the cli/go-gh repository using the GQL API.
func ExampleGQLClient_mutate() {
	client, err := api.DefaultGQLClient()
	if err != nil {
		log.Fatal(err)
	}
	var mutation struct {
		AddStar struct {
			Starrable struct {
				Repository struct {
					StargazerCount int
				} `graphql:"... on Repository"`
				Gist struct {
					StargazerCount int
				} `graphql:"... on Gist"`
			}
		} `graphql:"addStar(input: $input)"`
	}
	// Note that the shurcooL/githubv4 package has defined input structs generated from the
	// GraphQL schema that can be used instead of writing your own.
	type AddStarInput struct {
		StarrableID string `json:"starrableId"`
	}
	variables := map[string]interface{}{
		"input": AddStarInput{
			StarrableID: "R_kgDOF_MgQQ",
		},
	}
	err = client.Mutate("AddStar", &mutation, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mutation.AddStar.Starrable.Repository.StargazerCount)
}

// Query releases from cli/cli repository using GQL API with paginated results.
func ExampleGQLClient_pagination() {
	client, err := api.DefaultGQLClient()
	if err != nil {
		log.Fatal(err)
	}
	var query struct {
		Repository struct {
			Releases struct {
				Nodes []struct {
					Name string
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"releases(first: 30, after: $endCursor)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"owner":     graphql.String("cli"),
		"name":      graphql.String("cli"),
		"endCursor": (*graphql.String)(nil),
	}
	page := 1
	for {
		if err := client.Query("RepositoryReleases", &query, variables); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Page: %d\n", page)
		fmt.Println(query.Repository.Releases.Nodes)
		if !query.Repository.Releases.PageInfo.HasNextPage {
			break
		}
		variables["endCursor"] = graphql.String(query.Repository.Releases.PageInfo.EndCursor)
		page++
	}
}

// Get repository for the current directory.
func ExampleCurrent() {
	repo, err := repository.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s/%s/%s\n", repo.Host, repo.Owner, repo.Name)
}

// Print tabular data to a terminal or in machine-readable format for scripts.
func ExampleTablePrinter() {
	terminal := term.FromEnv()
	termWidth, _, _ := terminal.Size()
	t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)

	red := func(s string) string {
		return "\x1b[31m" + s + "\x1b[m"
	}

	// add a field that will render with color and will not be auto-truncated
	t.AddField("1", tableprinter.WithColor(red), tableprinter.WithTruncate(nil))
	t.AddField("hello")
	t.EndRow()
	t.AddField("2")
	t.AddField("world")
	t.EndRow()
	if err := t.Render(); err != nil {
		log.Fatal(err)
	}
}
