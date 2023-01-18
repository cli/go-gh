package gh_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	gh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcooL/githubv4"
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
func ExampleRESTClient_simple() {
	client, err := gh.RESTClient(nil)
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
func ExampleRESTClient_advanced() {
	opts := api.ClientOptions{
		Host:      "github.com",
		AuthToken: "xxxxxxxxxx", // Replace with valid auth token.
		Headers:   map[string]string{"Time-Zone": "America/Los_Angeles"},
		Log:       os.Stdout,
	}
	client, err := gh.RESTClient(&opts)
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
	client, err := gh.RESTClient(&opts)
	if err != nil {
		log.Fatal(err)
	}

	// URL to cli/cli release v2.14.2 checksums.txt
	assetURL := "repos/cli/cli/releases/assets/71589494"
	resp, err := client.Request("GET", assetURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	f, err := os.CreateTemp("", "*_checksums.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Asset downloaded to %s\n", f.Name())
}

// Query tags from cli/cli repository using GQL API.
func ExampleGQLClient_simple() {
	client, err := gh.GQLClient(nil)
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
func ExampleGQLClient_advanced() {
	opts := api.ClientOptions{
		EnableCache: true,
		Timeout:     5 * time.Second,
	}
	client, err := gh.GQLClient(&opts)
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
func ExampleGQLClient_Mutate_simple() {
	client, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	var m struct {
		AddStar struct {
			Starrable struct {
				Repository struct {
					Stargazers struct {
						Nodes []struct {
							Login string
						}
					} `graphql:"stargazers(first: $first)"`
				} `graphql:"... on Repository"`
				Gist struct {
					StargazerCount int
					Url            string
				} `graphql:"... on Gist"`
			}
		} `graphql:"addStar(input: $input)"`
	}
	variables := map[string]interface{}{
		"input": githubv4.AddStarInput{
			StarrableID: githubv4.NewString("R_kgDOF_MgQQ"),
		},
		"first": graphql.Int(10),
	}
	err = client.Mutate("AddStar", &m, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(m.AddStar.Starrable.Repository.Stargazers.Nodes[0].Login)
}

// Get repository for the current directory.
func ExampleCurrentRepository() {
	repo, err := gh.CurrentRepository()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s/%s/%s\n", repo.Host(), repo.Owner(), repo.Name())
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
