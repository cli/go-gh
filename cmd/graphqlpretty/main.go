package main

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/pkg/graphqlpretty"
)

func main() {
	if err := graphqlpretty.Format(os.Stdout, os.Stdin, "  ", true); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
