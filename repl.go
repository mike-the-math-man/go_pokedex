package main

import (
	"fmt"
	"os"
	"strings"
)

func cleanInput(text string) []string {
	var output []string
	text = strings.TrimSpace(text)
	words := strings.Split(text, " ")
	for _, word := range words {
		word = strings.ToLower(word)
		if word != "" {
			output = append(output, word)
		}
	}
	return output
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	defer os.Exit(0)
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var registry map[string]cliCommand

func get_Commands() map[string]cliCommand {
	if registry == nil {
		registry = map[string]cliCommand{
			"exit": {
				name:        "exit",
				description: "Exit the Pokedex",
				callback:    commandExit,
			},
			"help": {
				name:        "help",
				description: "Displays a help message",
				callback:    help,
			},
		}
	}
	return registry
}

func do_cliCommand(s string) {
	cli_return, ok := get_Commands()[s]
	if ok {
		cli_return.callback()
	} else {
		fmt.Println("Unknown command")
	}
}

func help() error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	fmt.Println()
	commands := get_Commands()
	for _, value := range commands {
		fmt.Printf("%s: %s\n", value.name, value.description)
	}
	return nil
}
