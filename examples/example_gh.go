package example

import (
	"fmt"
	"os"
	"time"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/shurcooL/githubv4"
)

// Execute `gh issue list -R cli/cli`, and print the output.
func ExampleExec() {
	args := []string{"issue", "list", "-R", "cli/cli"}
	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(stdOut.String())
	fmt.Println(stdErr.String())
}

// Get tags from cli/cli repository using REST API.
func ExampleRESTClient_simple() {
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	response := []struct{ Name string }{}
	err = client.Get("repos/cli/cli/tags", &response)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(response)
}

// Get tags from cli/cli repository using REST API.
// Specifying host, auth token, headers and logging to stdout.
func ExampleRESTClient_advanced() {
	opts := api.ClientOptions{
		Host:      "github.com",
		AuthToken: "xxxxxxxxxx", // Replace with valid auth token
		Headers:   map[string]string{"Time-Zone": "America/Los_Angeles"},
		Log:       os.Stdout,
	}
	client, err := gh.RESTClient(&opts)
	if err != nil {
		fmt.Println(err)
		return
	}
	response := []struct{ Name string }{}
	err = client.Get("repos/cli/cli/tags", &response)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(response)
}

// Query tags from cli/cli repository using GQL API.
func ExampleGQLClient_simple() {
	client, err := gh.GQLClient(nil)
	if err != nil {
		fmt.Println(err)
		return
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
		"refPrefix": githubv4.String("refs/tags/"),
		"last":      githubv4.Int(30),
		"owner":     githubv4.String("cli"),
		"name":      githubv4.String("cli"),
	}
	err = client.Query(&query, variables)
	if err != nil {
		fmt.Println(err)
		return
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
		fmt.Println(err)
		return
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
		"refPrefix": githubv4.String("refs/tags/"),
		"last":      githubv4.Int(30),
		"owner":     githubv4.String("cli"),
		"name":      githubv4.String("cli"),
	}
	err = client.Query(&query, variables)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(query)
}

// Get repository for the current directory.
func ExampleCurrentRepository() {
	repo, err := gh.CurrentRepository()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s/%s/%s\n", repo.Host(), repo.Owner(), repo.Name())
}
