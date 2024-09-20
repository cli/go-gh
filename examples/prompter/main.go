package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cli/go-gh/v2/pkg/prompter"
)

func main() {
	p := prompter.New(os.Stdin, os.Stdout, os.Stderr)

	// Demonstrating single-option select / dropdown prompts
	cuisines := []string{"Italian", "Greek", "Indian", "Japanese", "American"}
	favorite, err := p.Select("Favorite cuisine?", "Italian", cuisines)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Favorite cuisine: %s\n", cuisines[favorite])

	// Demonstrating multi-option select / dropdown prompts
	favorites, err := p.MultiSelect("Favorite cuisines?", []string{}, cuisines)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range favorites {
		fmt.Printf("Favorite cuisine: %s\n", cuisines[f])
	}

	// Demonstrating text input prompts
	text, err := p.Input("Favorite meal?", "Breakfast")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Favorite meal: %s\n", text)

	// Demonstrating password input prompts
	safeword, err := p.Password("Safe word?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Safe word: %s\n", safeword)

	// Demonstrating confirmation prompts
	confirmation, err := p.Confirm("Are you sure?", false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Confirmation: %t\n", confirmation)
}
