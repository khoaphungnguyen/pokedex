package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	"github.com/khoaphungnguyen/pokedex/cmd/pokecommand"
	"github.com/khoaphungnguyen/pokedex/internal/pokecache"
)

var commands map[string]pokecommand.CliCommand
var defaultURL = "https://pokeapi.co/api/v2/location-area/"
var pokemonURL = "https://pokeapi.co/api/v2/pokemon/"
var cache *pokecache.Cache
var locations *pokecommand.LocationArea

func init() {
	cache = pokecache.NewCache(5*time.Minute, 30*time.Minute)
	locations = &pokecommand.LocationArea{} 

	commands = map[string]pokecommand.CliCommand{
		"help": {
			Name:        "help",
			Description: "Displays a help message",
			Callback:    func(args []string) error{ return pokecommand.CommandHelp(commands) },
		},
		"exit": {
			Name:        "exit",
			Description: "Exit the Pokedex",
			Callback:    func(args []string) error { return pokecommand.CommandExit() },
		},
		"map": {
			Name:        "map",
			Description: "Displays the names of 20 location areas",
			// Use closure to capture additional parameters for CommandMap
			Callback: func(args []string) error{ return pokecommand.CommandMap(cache, locations, defaultURL) },
		},
		"mapb": {
			Name:        "mapb",
			Description: "Displays the previous 20 locations.",
			// Assume CommandBMap now correctly accepts a *LocationArea parameter
			Callback: func(args []string) error{ return pokecommand.CommandBMap(locations) },
		},
		"explore": {
			Name:        "explore",
			Description: "Explore a given area to list all Pokémon found there.",
			Callback:    func(args []string) error {
				if len(args) < 2 {
					fmt.Println("Please specify an area to explore.")
					return nil // Return nil error to indicate missing arguments but not a failure
				}
				areaName := args[1]
				return pokecommand.CommandExplore(areaName, cache, defaultURL)
			},
		},
		"catch": {
			Name:        "catch",
			Description: "Catch Pokemon and adds them to the user's Pokedex.",
			Callback:    func(args []string) error {
				if len(args) < 2 {
					fmt.Println("Please specify a Pokemon to catch.")
					return nil // Return nil error to indicate missing arguments but not a failure
				}
				pokemon := args[1]
				return pokecommand.CommandCatch(pokemon, cache, &pokecommand.Pokemon{}, pokemonURL)
			},
		},
		"inspect": {
			Name:        "inspect",
			Description: "Display the name, height, weight, stats and type(s) of the catched Pokemon.",
			Callback:    func(args []string) error {
				if len(args) < 2 {
					fmt.Println("Please specify a Pokemon to inspect.")
					return nil // Return nil error to indicate missing arguments but not a failure
				}
				pokemon := args[1]
				return pokecommand.CommandInspect(pokemon, cache, &pokecommand.Pokemon{})
			},
		},
		"pokedex": {
			Name:        "pokedex",
			Description: "Displays the names of catched Pokemons",
			// Use closure to capture additional parameters for CommandMap
			Callback: func(args []string) error{ return pokecommand.CommandPokedex(cache)},
		},
		
	}
	pokedexFile := "./pokedex.json"

	pokecommand.RegisterExitCallback(func() {
        // Perform necessary cleanup
        saveAndExit(pokedexFile)
    })
}

func saveAndExit(pokedexFile string) {
    fmt.Println("Saving Pokédex before exiting...")
    if err := cache.SavePokedex(pokedexFile); err != nil {
        fmt.Printf("Failed to save pokedex: %v\n", err)
    } else {
        fmt.Println("Pokédex saved successfully.")
    }
    os.Exit(0)
}

func main() {
	// Load Pokédex from file
	pokedexFile := "./pokedex.json"
	if err := cache.LoadPokedex(pokedexFile); err != nil {
		fmt.Printf("Failed to load pokedex: %v\n", err)
		return
	}

	// Setup signal handling to save Pokédex on program exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("Saving Pokédex before exiting...") 
		if err := cache.SavePokedex(pokedexFile); err != nil {
			fmt.Printf("Failed to save pokedex: %v\n", err)
		} else {
			fmt.Println("Pokédex saved successfully.")
		}
		os.Exit(1)
	}()

	fmt.Println("Simple Go REPL-Pokedex (type 'help' to explore menus or 'exit' to quit)")

	l, err := readline.NewEx(&readline.Config{
		Prompt:          "Pokedex> ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    nil,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		HistorySearchFold:   true,
	})
	if err != nil {
		fmt.Println("Error initializing readline:", err)
		return
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err != nil { // EOF, or Ctrl+C
			break
		}
		input := strings.TrimSpace(line)
		args := strings.Fields(strings.ToLower(input))

		if len(args) == 0 {
			continue
		}

		commandName := args[0]
		if command, exists := commands[commandName]; exists {
			err := command.Callback(args)
			if err != nil {
				fmt.Println("Error:", err)
			}
		} else {
			fmt.Println("Unknown command:", commandName)
		}
	}
}