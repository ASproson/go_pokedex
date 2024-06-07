package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	Previous string
	Next     string
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
	}
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
		Next: "https://pokeapi.co/api/v2/location?offset=0&limit=20",
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

		// Clean and process input
		command := cleanInput(input)

		// Check if command exists in map
		if cmd, exists := commands[command]; exists {
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
