package gh

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

// Execute 'gh issue list -R cli/cli', and print the output.
func ExampleExec() {
	args := []string{"issue", "list", "-R", "cli/cli"}
	stdOut, stdErr, err := Exec(args...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stdOut.String())
	fmt.Println(stdErr.String())
}

// Get tags from cli/cli repository using REST API.
func ExampleRESTClient_simple() {
	client, err := RESTClient(nil)
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
	client, err := RESTClient(&opts)
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

// Query tags from cli/cli repository using GQL API.
func ExampleGQLClient_simple() {
	client, err := GQLClient(nil)
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
	client, err := GQLClient(&opts)
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

// Get repository for the current directory.
func ExampleCurrentRepository() {
	repo, err := CurrentRepository()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s/%s/%s\n", repo.Host(), repo.Owner(), repo.Name())
}

// Use SkipResolution ClientOption to change a http.Client into a api.RESTClient.
func ExampleHTTPClient_skipResolution() {
	host := "github.com"
	httpOpts := api.ClientOptions{
		Host: host,
	}
	httpClient, err := HTTPClient(&httpOpts)
	if err != nil {
		log.Fatal(err)
	}
	// Use SkipResolution as our http.Client does the handling of
	// options and headers.
	restOpts := api.ClientOptions{
		SkipResolution: true,
		Host:           host,
		Transport:      httpClient.Transport,
	}
	restClient, err := RESTClient(&restOpts)
	if err != nil {
		log.Fatal(err)
	}
	response := []struct{ Name string }{}
	err = restClient.Get("repos/cli/cli/tags", &response)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}
