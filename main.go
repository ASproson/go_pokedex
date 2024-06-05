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

func commandHelp() error {
	fmt.Println()
	fmt.Println("Welcome to the Pokédex!")
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("	help: Displays a help message")
	fmt.Println("	exit: Exit the Pokédex")
	fmt.Println()
	return nil
}

func commandExit() error {
	fmt.Println()
	fmt.Println("Exiting Pokédex...")
	os.Exit(0)
	return nil
}

func main() {

	commands := map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exits the Pokédex",
			callback:    commandExit,
		},
	}

	// REPL loop
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Pokédex > ")

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Remove white spaces
		command := strings.TrimSpace(input)

		// Check if command exists in map
		if cmd, exists := commands[command]; exists {
			// Execute command
			if err := cmd.callback(); err != nil {
				fmt.Println("Error executing command:", err)
			}
		} else {
			fmt.Println()
			fmt.Println("Unknown command:", command)
			fmt.Println()
		}
	}
}
