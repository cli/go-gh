package main

import (
	"fmt"

	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/auth"
	"github.com/cli/go-gh/pkg/config"
	"github.com/cli/go-gh/pkg/git"
)

func main() {
	remotes, err := git.Remotes()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(remotes) == 0 {
		fmt.Println("unable to determine current repo")
		return
	}
	remote := remotes[0]
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(err)
		return
	}
	host := cfg.Host()
	token, err := auth.Token(host)
	if err != nil {
		fmt.Println(err)
		return
	}
	opts := api.ClientOptions{AuthToken: token}
	client := api.NewRESTClient(host, opts)
	response := []struct{ Name string }{}
	path := fmt.Sprintf("repos/%s/%s/tags", remote.Owner, remote.Repo)
	err = client.Get(path, &response)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(response)
	return
}
