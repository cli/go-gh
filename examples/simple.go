package main

/* func main() {
	// Retreive a sorted list of git remotes for current directory.
	// Will return an error if git is not found in path, or if current directory is not a git directory.
	remotes, err := git.Remotes()
	if err != nil {
		fmt.Println(err)
		return
	}

	// If no remotes are found we are unable to determine corresponding repository for current directory.
	if len(remotes) == 0 {
		fmt.Println("unable to determine current repository")
		return
	}
	remote := remotes[0]

	// Load gh configuration file.
	// Will return a default configuration if one does not exist.
	// Will return an error if trouble reading configuration file.
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get target host for operations.
	// Takes into account configuration and environment variables.
	host := cfg.Host()

	// Retrieve authentication token for host.
	// Will return an error if no token found.
	// Takes into account configuration and environment variables.
	token, err := auth.Token(host)
	if err != nil {
		fmt.Println(err)
		return
	}

	opts := api.ClientOptions{AuthToken: token}

	client := api.NewRESTClient(host, opts)

	path := fmt.Sprintf("repos/%s/%s/tags", remote.Owner, remote.Repo)

	response := []struct{ Name string }{}

	err = client.Get(path, &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(response)
} */
