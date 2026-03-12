package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mike-the-math-man/go_pokedex/internal"
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

func commandExit(c *config, cache *internal.Cache) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	defer os.Exit(0)
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, *internal.Cache) error
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
			"map": {
				name:        "map",
				description: "Displays the names of 20 location areas in the Pokemon world",
				callback:    map_func,
			},
			"mapb": {
				name:        "mapb",
				description: "Displays the names of the previous 20 location areas in the Pokemon world",
				callback:    map_func_back,
			},
		}
	}
	return registry
}

func do_cliCommand(s string, c *config, cached_data *internal.Cache) {
	cli_return, ok := get_Commands()[s]
	if ok {
		cli_return.callback(c, cached_data)
	} else {
		fmt.Println("Unknown command")
	}
}

func help(c *config, cache *internal.Cache) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	fmt.Println()
	commands := get_Commands()
	for _, value := range commands {
		fmt.Printf("%s: %s\n", value.name, value.description)
	}
	return nil
}

func map_func(c *config, cached_data *internal.Cache) error {
	map_locations, _ := get_map_data(c, true, cached_data)
	for i := 0; i < 20; i++ {
		if i >= len(map_locations.Results) {
			return fmt.Errorf("Slice index %d out of bounds (slice length = %d)", i, len(map_locations.Results))
		}
		fmt.Println(map_locations.Results[i].Name)
	}
	return nil
}

type config struct {
	Next     string
	Previous string
}

type location struct {
	Name string
	Url  string
}

type poke_api_data struct {
	Count    int
	Next     string
	Previous string
	Results  []location
}

func get_map_data(c *config, s bool, cached_data *internal.Cache) (poke_api_data, error) {
	var url string
	if s {
		url = c.Next
	} else {
		url = c.Previous
	}
	if url == "" {
		fmt.Println("you're on the first page")
		return poke_api_data{}, nil
	}
	cache_map, ok := cached_data.Get(url)
	if ok {
		var decoded_cache_map poke_api_data
		err := json.Unmarshal(cache_map, &decoded_cache_map)
		if err != nil {
			fmt.Println("error decoding Request")
			return poke_api_data{}, err
		}
		c.Previous = decoded_cache_map.Previous
		c.Next = decoded_cache_map.Next
		return decoded_cache_map, nil
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("error creating NewRequest")
		return poke_api_data{}, err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error client Do request")
		return poke_api_data{}, err
	}
	defer res.Body.Close()

	var decoded_response poke_api_data
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&decoded_response)
	if err != nil {
		fmt.Println("error decoding Request")
		return poke_api_data{}, err
	}
	c.Previous = decoded_response.Previous
	c.Next = decoded_response.Next
	cache_write, err := json.Marshal(decoded_response)
	if err != nil {
		fmt.Println("error encoding Request")
		return poke_api_data{}, err
	}

	cached_data.Add(url, cache_write)
	return decoded_response, nil
}

func map_func_back(c *config, cached_data *internal.Cache) error {
	map_locations, _ := get_map_data(c, false, cached_data)
	for i := 0; i < 20; i++ {
		if i >= len(map_locations.Results) {
			return fmt.Errorf("Slice index %d out of bounds (slice length = %d)", i, len(map_locations.Results))
		}
		fmt.Println(map_locations.Results[i].Name)
	}
	return nil
}
