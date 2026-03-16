package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"

	"math/rand"

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

func commandExit(s string, c *config, cache *internal.Cache) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	defer os.Exit(0)
	return nil
}

type cliCommand struct {
	name        string
	description string
	callback    func(string, *config, *internal.Cache) error
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
			"explore": {
				name:        "explore",
				description: "Explores a location - takes name of location as input",
				callback:    explore,
			},
			"catch": {
				name:        "catch",
				description: "Attempts to catch a Pokemon",
				callback:    catch,
			},
			"inspect": {
				name:        "inspect",
				description: "Displays information about caught pokemon",
				callback:    inspect,
			},
			"pokedex": {
				name:        "pokedex",
				description: "List the pokemon that player has caught",
				callback:    pokedex,
			},
		}
	}
	return registry
}

func do_cliCommand(s string, param string, c *config, cached_data *internal.Cache) {
	cli_return, ok := get_Commands()[s]
	if ok {
		cli_return.callback(param, c, cached_data)
	} else {
		fmt.Println("Unknown command")
	}
}

func help(s string, c *config, cache *internal.Cache) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	fmt.Println()
	commands := get_Commands()
	for _, value := range commands {
		fmt.Printf("%s: %s\n", value.name, value.description)
	}
	return nil
}

func map_func(s string, c *config, cached_data *internal.Cache) error {
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
	Pokedex  map[string]Pokemon
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

func map_func_back(s string, c *config, cached_data *internal.Cache) error {
	map_locations, _ := get_map_data(c, false, cached_data)
	for i := 0; i < 20; i++ {
		if i >= len(map_locations.Results) {
			return fmt.Errorf("Slice index %d out of bounds (slice length = %d)", i, len(map_locations.Results))
		}
		fmt.Println(map_locations.Results[i].Name)
	}
	return nil
}

type Resource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ExploreData struct {
	EncounterMethodRates []struct {
		EncounterMethod Resource `json:"encounter_method"`
		VersionDetails  []struct {
			Rate    int      `json:"rate"`
			Version Resource `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int      `json:"game_index"`
	ID        int      `json:"id"`
	Location  Resource `json:"location"`
	Name      string   `json:"name"`
	Names     []struct {
		Language Resource `json:"language"`
		Name     string   `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon        Resource `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int        `json:"chance"`
				ConditionValues []Resource `json:"condition_values"`
				MaxLevel        int
				Method          Resource `json:"method"`
				MinLevel        int
			} `json:"encounter_details"`
			MaxChance int      `json:"max_chance"`
			Version   Resource `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

func explore(location string, c *config, cached_data *internal.Cache) error {
	//scanner := bufio.NewScanner(os.Stdin)
	//fmt.Print("Location > ")
	//scanner.Scan()

	if location == "" {
		//fmt.Println("Please enter location after explore")
		return nil
	}
	fmt.Printf("Exploring %s...\n", location)
	fmt.Println("Found Pokemon:")
	url := "https://pokeapi.co/api/v2/location-area/" + location //cleanInput(scanner.Text())[0]

	cache_pokemans, ok := cached_data.Get(url)
	if ok {
		var decoded_cache_pokemans ExploreData
		err := json.Unmarshal(cache_pokemans, &decoded_cache_pokemans)
		if err != nil {
			fmt.Println("error decoding Request")
			return err
		}
		for _, j := range decoded_cache_pokemans.PokemonEncounters {
			fmt.Printf(" -%s\n", j.Pokemon.Name)
		}
		return nil
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("error creating NewRequest")
		return err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error client Do request")
		return err
	}
	defer res.Body.Close()

	var decoded_explore_data ExploreData
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&decoded_explore_data)
	if err != nil {
		fmt.Println("error decoding Request")
		return err
	}
	for _, j := range decoded_explore_data.PokemonEncounters {
		fmt.Printf(" -%s\n", j.Pokemon.Name)
	}

	cache_write, err := json.Marshal(decoded_explore_data)
	if err != nil {
		fmt.Println("error encoding Request")
		return err
	}

	cached_data.Add(url, cache_write)

	return nil
}

type Pokemon struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Forms          []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndices []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	Height                 int           `json:"height"`
	HeldItems              []interface{} `json:"held_items"`
	ID                     int           `json:"id"`
	IsDefault              bool          `json:"is_default"`
	LocationAreaEncounters string        `json:"location_area_encounters"`
	Moves                  []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
			VersionGroup struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Name      string        `json:"name"`
	Order     int           `json:"order"`
	PastTypes []interface{} `json:"past_types"`
	Species   struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string      `json:"back_default"`
		BackFemale       interface{} `json:"back_female"`
		BackShiny        string      `json:"back_shiny"`
		BackShinyFemale  interface{} `json:"back_shiny_female"`
		FrontDefault     string      `json:"front_default"`
		FrontFemale      interface{} `json:"front_female"`
		FrontShiny       string      `json:"front_shiny"`
		FrontShinyFemale interface{} `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault string      `json:"front_default"`
				FrontFemale  interface{} `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string      `json:"front_default"`
				FrontFemale      interface{} `json:"front_female"`
				FrontShiny       string      `json:"front_shiny"`
				FrontShinyFemale interface{} `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           string `json:"back_default"`
					BackShiny             string `json:"back_shiny"`
					BackShinyTransparent  string `json:"back_shiny_transparent"`
					BackTransparent       string `json:"back_transparent"`
					FrontDefault          string `json:"front_default"`
					FrontShiny            string `json:"front_shiny"`
					FrontShinyTransparent string `json:"front_shiny_transparent"`
					FrontTransparent      string `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      string      `json:"back_default"`
						BackFemale       interface{} `json:"back_female"`
						BackShiny        string      `json:"back_shiny"`
						BackShinyFemale  interface{} `json:"back_shiny_female"`
						FrontDefault     string      `json:"front_default"`
						FrontFemale      interface{} `json:"front_female"`
						FrontShiny       string      `json:"front_shiny"`
						FrontShinyFemale interface{} `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string      `json:"back_default"`
					BackFemale       interface{} `json:"back_female"`
					BackShiny        string      `json:"back_shiny"`
					BackShinyFemale  interface{} `json:"back_shiny_female"`
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault string      `json:"front_default"`
					FrontFemale  interface{} `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     string      `json:"front_default"`
					FrontFemale      interface{} `json:"front_female"`
					FrontShiny       string      `json:"front_shiny"`
					FrontShinyFemale interface{} `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				Icons struct {
					FrontDefault string      `json:"front_default"`
					FrontFemale  interface{} `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
}

func catch(pokemon string, c *config, cache *internal.Cache) error {

	if pokemon == "" {
		return nil
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon)

	url := "https://pokeapi.co/api/v2/pokemon/" + pokemon

	cache_pokemans, ok := cache.Get(url)
	if ok {
		var decoded_cache_pokemans Pokemon
		err := json.Unmarshal(cache_pokemans, &decoded_cache_pokemans)
		if err != nil {
			fmt.Println("error decoding Request")
			return err
		}
		catch_role := rand.Float64()

		catch_difficulty := math.Log10(float64(decoded_cache_pokemans.BaseExperience)) / float64(2.81)

		if catch_role >= float64(catch_difficulty) {
			//fmt.Println(catch_role)
			fmt.Printf("%s was caught!\n", pokemon)
			c.Pokedex[pokemon] = decoded_cache_pokemans
		} else {
			fmt.Printf("%s escaped!\n", pokemon)
		}
		return nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("error creating NewRequest")
		return err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error client Do request")
		return err
	}
	defer res.Body.Close()

	var decoded_pokemon_data Pokemon
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&decoded_pokemon_data)
	if err != nil {
		fmt.Println("error decoding Request")
		return err
	}

	catch_role := rand.Float64()

	catch_difficulty := math.Log10(float64(decoded_pokemon_data.BaseExperience)) / float64(2.81)

	if catch_role >= float64(catch_difficulty) {
		//fmt.Println(catch_role)
		fmt.Printf("%s was caught!\n", pokemon)
		c.Pokedex[pokemon] = decoded_pokemon_data
	} else {
		fmt.Printf("%s escaped!\n", pokemon)
	}

	cache_write, err := json.Marshal(decoded_pokemon_data)
	if err != nil {
		fmt.Println("error encoding Request")
		return err
	}

	cache.Add(url, cache_write)

	return nil
}

func inspect(pokemon string, c *config, cache *internal.Cache) error {
	inspected_pokemon, ok := c.Pokedex[pokemon]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return fmt.Errorf("can't display for uncaught pokemon")
	}
	fmt.Printf("Name: %s\n", inspected_pokemon.Name)
	fmt.Printf("Height: %d\n", inspected_pokemon.Height)
	fmt.Printf("Weight: %d\n", inspected_pokemon.Weight)
	//fmt.Println(inspected_pokemon)

	fmt.Println("Stats:")
	for _, stats := range inspected_pokemon.Stats {
		fmt.Printf(" -%s: %d\n", stats.Stat.Name, stats.BaseStat)
	}
	fmt.Println("Types:")
	for _, types := range inspected_pokemon.Types {
		fmt.Printf(" -%s\n", types.Type.Name)
	}
	return nil
}

func pokedex(pokemon string, c *config, cache *internal.Cache) error {
	fmt.Println(("Your Pokedex:"))
	for _, value := range c.Pokedex {
		fmt.Printf(" -%s\n", value.Name)
	}
	return nil
}
