package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/mike-the-math-man/go_pokedex/internal"
)

func main() {
	cached_data := internal.NewCache(5 * time.Second)
	scanner := bufio.NewScanner(os.Stdin)
	current_config := config{
		Next:     "https://pokeapi.co/api/v2/location-area/",
		Previous: "",
		Pokedex:  map[string]Pokemon{},
	}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		clean_input := cleanInput(scanner.Text())
		if len(clean_input) > 1 {
			command := clean_input[0]
			paramater := clean_input[1]
			do_cliCommand(command, paramater, &current_config, cached_data)
		}
		if len(clean_input) > 0 {
			command := clean_input[0]
			do_cliCommand(command, "", &current_config, cached_data)
		}
		//fmt.Println("Please enter a command - or type help")
		//clean_text := cleanInput(scanner.Text())
		//command := clean_text[0]
		//fmt.Printf("Your command was: %s\n", clean_text[0])
	}
}
