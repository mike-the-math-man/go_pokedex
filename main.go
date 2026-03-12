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
	}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		do_cliCommand(cleanInput(scanner.Text())[0], &current_config, cached_data)
		//clean_text := cleanInput(scanner.Text())
		//command := clean_text[0]
		//fmt.Printf("Your command was: %s\n", clean_text[0])
	}
}
