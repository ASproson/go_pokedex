package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
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
func commandMap(c *config) error {
	if c.Next == "" {
		fmt.Println("\nNo more locations to display")
		return nil
	}

	pokemonLocales, err := fetchAndPrintLocations(c.Next)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Previous = c.Next
	c.Next = pokemonLocales.Next
	return nil
}

// Prints the previous 20 locations, errors when used on first page
func commandMapBack(c *config) error {
	if c.Previous == "" {
		fmt.Println("\nNo previous locations to display")
		return nil
	}

	pokemonLocales, err := fetchAndPrintLocations(c.Previous)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.Next = pokemonLocales.Next

	if pokemonLocales.Previous == nil {
		c.Previous = ""
	} else {
		c.Previous = pokemonLocales.Previous.(string)
	}

	return nil
}

func fetchAndPrintLocations(url string) (PokemonLocales, error) {
	res, err := http.Get(url)
	if err != nil {
		return PokemonLocales{}, fmt.Errorf("failed to fetch locations: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return PokemonLocales{}, fmt.Errorf("failed to read response body: %v", err)
	}

	if res.StatusCode > 299 {
		return PokemonLocales{}, fmt.Errorf("response failed with status code: %d, and body: %s", res.StatusCode, body)
	}

	var pokemonLocales PokemonLocales
	err = json.Unmarshal(body, &pokemonLocales)
	if err != nil {
		return PokemonLocales{}, fmt.Errorf("failed to parse JSON: %v", err)
	}

	for _, locale := range pokemonLocales.Results {
		fmt.Println(locale.Name)
	}

	return pokemonLocales, nil
}

// Trims whitespace and converts input to lowercase
func cleanInput(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

// Prints a help message listing all available commands
func commandHelp(c *config) error {
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
func commandExit(c *config) error {
	fmt.Println("\nExiting Pokédex...")
	os.Exit(0)
	return nil
}

func main() {
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
			if err := cmd.callback(conf); err != nil {
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
