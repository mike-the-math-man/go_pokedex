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

type Ability struct {
	resource  Resource
	is_hidden bool
	slot      int
}

type Pokemon struct {
	Abilities       []Ability `json:"abilities"`
	Base_experience int       `json:"base_experience"`
	Cries           struct {
		Latest string `json:"latest"`
		Legacy string `json:"legacy"`
	} `json:"cries"`
	Forms        []Resource `json:"forms"`
	Game_indeces []struct {
		game_index int
		version    Resource
	} `json:"game_indeces"`
	Height     int `json:"height"`
	Held_items []struct {
		item            Resource
		version_details []struct {
			rarity  int
			version Resource
		}
	} `json:"held_items"`
	Id                       int    `json:"id"`
	Is_default               bool   `json:"is_default"`
	Location_area_encounters string `json:"location_area_encounters"`
	Moves                    []struct {
		move                  Resource
		version_group_details []struct {
			level_learned_at  int
			move_learn_method Resource
			order             int
			version_group     Resource
		}
	} `json:"moves"`
	Name           string `json:"name"`
	Order          int    `json:"order"`
	Past_abilities []struct {
		abilities []struct {
			ability   string
			is_hidden bool
			slot      int
		}
		generation Resource
	} `json:"past_abilities"`
	Past_stats []struct {
		generation Resource
		stats      []struct {
			base_stat int
			effort    int
			stat      Resource
		}
	} `json:"past_stats"`
	Past_types []struct {
		generation Resource
		types      []struct {
			slot int
			Type Resource `jsin:"type"`
		}
	} `json:"past_types"`
	Species Resource `json:"species"`
	Sprites struct {
		back_default       string
		back_female        string
		back_shiny         string
		back_shiny_female  string
		front_default      string
		front_female       string
		front_shiny        string
		front_shiny_female string
		other              struct {
			dream_world struct {
				front_default string
				front_female  string
			}
			home struct {
				front_default      string
				front_female       string
				front_shiny        string
				front_shiny_female string
			}
			official_artwork struct {
				front_default string
				front_shiny   string
			}
			showdown struct {
				back_default       string
				back_female        string
				back_shiny         string
				back_shiny_female  string
				front_default      string
				front_female       string
				front_shiny        string
				front_shiny_female string
			}
		}
		versions struct {
			generation_i struct {
				red_blue struct {
					back_default  string
					back_gray     string
					front_default string
					front_gray    string
				}
				yellow struct {
					back_default  string
					back_gray     string
					front_default string
					front_gray    string
				}
			}
			generation_ii struct {
				crystal struct {
					back_default  string
					back_shiny    string
					front_default string
					front_shiny   string
				}
				gold struct {
					back_default  string
					back_shiny    string
					front_default string
					front_shiny   string
				}
				silver struct {
					back_default  string
					back_shiny    string
					front_default string
					front_shiny   string
				}
			}
			generation_iii struct {
				emerald struct {
					front_default string
					front_shiny   string
				}
				firered_leafgreen struct {
					back_default  string
					back_shiny    string
					front_default string
					front_shiny   string
				}
				ruby_sapphire struct {
					back_default  string
					back_shiny    string
					front_default string
					front_shiny   string
				}
			}
			generation_iv struct {
				diamond_pearl struct {
					back_default       string
					back_female        string
					back_shiny         string
					back_shiny_female  string
					front_default      string
					front_female       string
					front_shiny        string
					front_shiny_female string
				}
				heartgold_silver struct {
					back_default       string
					back_female        string
					back_shiny         string
					back_shiny_female  string
					front_default      string
					front_female       string
					front_shiny        string
					front_shiny_female string
				}
				platinum struct {
					back_default       string
					back_female        string
					back_shiny         string
					back_shiny_female  string
					front_default      string
					front_female       string
					front_shiny        string
					front_shiny_female string
				}
			}
			generation_v struct {
				black_white struct {
					animated struct {
						back_default       string
						back_female        string
						back_shiny         string
						back_shiny_female  string
						front_default      string
						front_female       string
						front_shiny        string
						front_shiny_female string
					}
					back_default       string
					back_female        string
					back_shiny         string
					back_shiny_female  string
					front_default      string
					front_female       string
					front_shiny        string
					front_shiny_female string
				}
			}
			generation_vi struct {
				omegaruby_alphasapphire struct {
					front_default      string
					front_female       string
					front_shiny        string
					front_shiny_female string
				}
				x_y struct {
					front_default      string
					front_female       string
					front_shiny        string
					front_shiny_female string
				}
			}
			generation_vii struct {
				icons struct {
					front_default string
					front_shiny   string
				}
				ultra_sun_ultra_moon struct {
					front_default      string
					front_female       string
					front_shiny        string
					front_shiny_female string
				}
			}
			generation_viii struct {
				icons struct {
					front_default string
					front_female  string
				}
			}
		}
	} `json:"sprites"`
	Stats []struct {
		base_stat int
		effort    int
		stat      Resource
	} `json:"stats"`
	Types []struct {
		slot int
		Type Resource `json:"type"`
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

		catch_difficulty := math.Log10(float64(decoded_cache_pokemans.Base_experience)) / float64(2.81)

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

	catch_difficulty := math.Log10(float64(decoded_pokemon_data.Base_experience)) / float64(2.81)

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
