package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
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
	}
}

// Trims whitespace and converts input to lowercase
func cleanInput(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

// Prints a help message listing all available commands
func commandHelp() error {
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
func commandExit() error {
	fmt.Println("\nExiting Pokédex...")
	os.Exit(0)
	return nil
}

func main() {
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
			if err := cmd.callback(); err != nil {
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
