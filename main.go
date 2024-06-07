package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/asproson/go_pokedex/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, *pokecache.Cache) error
}

type config struct {
	Previous   string
	Next       string
	CurrentArg string
	Pokedex    *Pokedex
}

type PokemonLocales struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationAreaResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokedex struct {
	Pokemon map[string]Pokemon
}

type Pokemon struct {
	Name           string         `json:"name"`
	BaseExperience int            `json:"base_experience"`
	Height         int            `json:"height"`
	Weight         int            `json:"weight"`
	Stats          []PokemonStats `json:"stats"`
	Types          []PokemonTypes `json:"types"`
}

type PokemonStats struct {
	Stat struct {
		Name string `json:"name"`
	} `json:"stat"`
	BaseStat int `json:"base_stat"`
}

type PokemonTypes struct {
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

// Returns a map of available commands
func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exits the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Shows the first 20 locations in the Pokémon world, each subsequent call shows the next 20",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Shows the previous 20 locations in the Pokémon world, each subsequent call shows the previous 20",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Shows all available Pokémon on the passed route",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch the named Pokémon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Use the Pokédex to inspect your caught Pokémon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Use the Pokédex to see all the Pokémon you have caught",
			callback:    commandPokedex,
		},
	}
}

func commandPokedex(c *config, cache *pokecache.Cache) error {
	if len(c.Pokedex.Pokemon) == 0 {
		fmt.Println("No Pokémon have been caught")
		return nil
	}

	fmt.Println("Your Pokédex:")

	for _, pokemon := range c.Pokedex.Pokemon {
		fmt.Printf("- %s\n", pokemon.Name)
	}
	return nil
}

func commandInspect(c *config, cache *pokecache.Cache) error {
	pokemonName := c.CurrentArg

	if pokemonName == "" {
		fmt.Println("Pokémon name is required to inspect")
		return nil
	}

	// Check if Pokémon is in Pokédex
	pokemon, exists := c.Pokedex.Pokemon[pokemonName]
	if !exists {
		fmt.Printf("You have not yet caught %s\n", pokemonName)
		return nil
	}

	// Pokémon exists, print details
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, pokemonType := range pokemon.Types {
		fmt.Printf("  - %s\n", pokemonType.Type.Name)
	}

	return nil
}

func commandCatch(c *config, cache *pokecache.Cache) error {
	pokemonToCatch := c.CurrentArg

	if pokemonToCatch == "" {
		fmt.Println("Pokémon name is required to catch")
		return nil
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", pokemonToCatch)

	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch Pokémon: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Decode JSON response into Pokemon struct
	var fetchedPokemon Pokemon
	if err := json.Unmarshal(body, &fetchedPokemon); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	fmt.Printf("...%s was found! Throwing Pokéball!\n", fetchedPokemon.Name)

	// Simulate catching experience
	for i := 0; i < 3; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("...")
	}

	// The higher the base experience, the harder it is to catch
	catchChance := rand.Intn(100) // Random number between 0-99

	if catchChance < 50 {
		fmt.Printf("%s was caught!\n", fetchedPokemon.Name)
		c.Pokedex.Pokemon[pokemonToCatch] = fetchedPokemon
	} else {
		fmt.Printf("%s escaped!\n", fetchedPokemon.Name)
	}

	return nil

}

func commandExplore(c *config, cache *pokecache.Cache) error {
	locationName := c.CurrentArg

	if locationName == "" {
		fmt.Println("Location area name is required for explore command")
		return nil
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", locationName)

	// Check if locale detail is in cache
	if val, found := cache.Get(url); found {
		fmt.Println(">>>>> Using cached data <<<<<")
		return printPokemonNames(val)
	}

	res, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("failed to fetch location: %v", locationName)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read resposne body: %v", err)
	}

	// Add response to the cache
	cache.Add(url, body)

	return printPokemonNames(body)
}

func printPokemonNames(body []byte) error {
	var data LocationAreaResponse

	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	fmt.Println("Found Pokémon:")
	for _, encounter := range data.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

// Prints the next 20 locations
func commandMap(c *config, cache *pokecache.Cache) error {
	if c.Next == "" {
		fmt.Println("\nNo more locations to display")
		return nil
	}

	err := fetchAndPrintLocations(c, cache)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Prints the previous 20 locations, errors when used on first page
func commandMapBack(c *config, cache *pokecache.Cache) error {
	if c.Previous == "" {
		fmt.Println("\nNo previous locations to display")
		return nil
	}

	// Swap c.Next and c.Previous before fetching the locations
	c.Next, c.Previous = c.Previous, c.Next

	err := fetchAndPrintLocations(c, cache)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func fetchAndPrintLocations(c *config, cache *pokecache.Cache) error {
	url := c.Next

	// Check if the response for this URL is in the cache
	if val, found := cache.Get(url); found {
		fmt.Println(">>>>> Using cached data <<<<<")
		return updateConfigAndPrintLocations(c, val)
	}

	// Make the network request since it's not in the cache
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch locations: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Add the response to the cache
	cache.Add(url, body)

	// Print the locations and update the config
	return updateConfigAndPrintLocations(c, body)
}

func updateConfigAndPrintLocations(c *config, body []byte) error {
	var data PokemonLocales

	// Unmarshal the response body
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	for _, result := range data.Results {
		fmt.Println(result.Name)
	}

	c.Next = data.Next
	if data.Previous == nil {
		c.Previous = ""
	} else {
		c.Previous = data.Previous.(string)
	}

	return nil
}

// Trims whitespace and converts input to lowercase
func cleanInput(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

// Prints a help message listing all available commands
func commandHelp(c *config, cache *pokecache.Cache) error {
	fmt.Println("\nWelcome to the Pokédex!")
	fmt.Println("Available commands:")

	commands := getCommands()
	for _, value := range commands {
		fmt.Printf("%s: %s\n", value.name, value.description)
	}
	fmt.Println()
	return nil
}

// Exit the application
func commandExit(c *config, cache *pokecache.Cache) error {
	fmt.Println("\nExiting Pokédex...")
	os.Exit(0)
	return nil
}

func main() {
	// Initial cache with cleanup interval of 5 minutes
	cache := pokecache.NewCache(5 * time.Minute)

	// Initialize config with inital url
	conf := &config{
		Next:    "https://pokeapi.co/api/v2/location?offset=0&limit=20",
		Pokedex: &Pokedex{Pokemon: make(map[string]Pokemon)},
	}

	commands := getCommands()
	// Collect user input
	scanner := bufio.NewScanner(os.Stdin)

	// REPL loop
	for {
		fmt.Print("Pokédex > ")
		// Read user input
		if !scanner.Scan() {
			fmt.Println("Error reading input or end of input detected")
			break
		}
		input := scanner.Text()
		cleanedInput := cleanInput(input)

		parts := strings.Fields(cleanedInput)
		if len(parts) == 0 {
			continue // skip empty input
		}

		// Clean and process input
		command := parts[0]

		var arg string
		if len(parts) > 1 {
			arg = parts[1]
		}

		// Check if command exists in map
		if cmd, exists := commands[command]; exists {
			conf.CurrentArg = arg
			// Execute the command's callback function
			if err := cmd.callback(conf, cache); err != nil {
				fmt.Println("Error executing command:", err)
			}
		} else {
			fmt.Println("Unknown command:", command)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error with input:", err)
	}
}
