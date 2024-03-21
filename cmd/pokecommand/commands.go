package pokecommand

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/khoaphungnguyen/pokedex/internal/pokecache"
)

var onExitCallback func()

type CliCommand struct {
	Name        string
	Description string
	Callback    func(args []string	) error
}

type LocationArea struct {
	Count    int `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationPokemonDetails struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
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
	Height    int `json:"height"`
	ID                     int    `json:"id"`
	Name          string `json:"name"`
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


func CommandHelp(commands map[string]CliCommand) error {
	fmt.Println("\nWelcome to the Pokedex!")
	fmt.Println("Usage:")
	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.Name, cmd.Description)
	}
	return nil
}

func CommandExit() error {
    fmt.Println("Exiting REPL...")
    if onExitCallback != nil {
        onExitCallback()
    }
    return nil
}


func CommandMap(cache *pokecache.Cache, locations *LocationArea, defaultURL string) error {
	url := locations.Next
	if url == "" {
		url = defaultURL
	}

	// Check if the data is in the cache
	if data, found := cache.Get(url); found {
		// Data is found in the cache, unmarshal and use it
		if err := json.Unmarshal(data, locations); err != nil {
			fmt.Println("Error decoding JSON from cache:", err)
			return err
		}
	} else {
		// Data not found in cache, proceed with the HTTP request
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error fetching data:", err)
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body) // Use io.ReadAll here
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}

		// Unmarshal the JSON data into the struct
		if err := json.Unmarshal(body, locations); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return err
		}

		// Store the fetched data in the cache for future use
		cache.Set(url, body, 30*time.Minute) // Customize the expiration as needed
	}

	// Print the result 
	for _, r := range locations.Results {
		fmt.Printf("%s\n", r.Name)
	}

	return nil
}


func CommandBMap(locations *LocationArea) error {
	url := locations.Previous
	if url == "" {
		fmt.Println("No previous maps available.")
		return nil
	}

	// Make GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data", err)
		return nil
	}
	defer resp.Body.Close()

	// Read the body of the response
	body, err := io.ReadAll(resp.Body)
	if err != nil{
		fmt.Println("Error reading response body:", err)
		return nil
	}

	//unmarshal the JSON data into the struct
	err = json.Unmarshal(body, &locations)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil
	}

	// Print the result 
	for _, r := range locations.Results {
		fmt.Printf("%s\n", r.Name)
	}

	return nil
}

func CommandExplore(areaName string, cache *pokecache.Cache, defaultURL string) error {
    fmt.Printf("Exploring %s...\n", areaName)
    url := defaultURL + areaName

    // Check cache first
    if data, found := cache.Get(url); found {
        var locationPokemonDetails LocationPokemonDetails
        if err := json.Unmarshal(data, &locationPokemonDetails); err != nil {
            fmt.Println("Error decoding JSON from cache:", err)
            return err
        }
        printPokemonNames(locationPokemonDetails)
    } else {
        // Data not found in cache, proceed with HTTP request
        resp, err := http.Get(url)
        if err != nil {
            fmt.Println("Error fetching data:", err)
            return err
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            fmt.Println("Error reading response body:", err)
            return err
        }

        var locationPokemonDetails LocationPokemonDetails
        if err := json.Unmarshal(body, &locationPokemonDetails); err != nil {
            fmt.Println("Error decoding JSON:", err)
            return err
        }

        // Update cache
        cache.Set(url, body, 30*time.Minute)

        printPokemonNames(locationPokemonDetails)
    }

    return nil
}

func printPokemonNames(locationPokemonDetails LocationPokemonDetails) {
    var wg sync.WaitGroup

for _, encounter := range locationPokemonDetails.PokemonEncounters {
    wg.Add(1)
    // Capture the current value of encounter in the loop.
    encounterCopy := encounter
    go func() {
        defer wg.Done()
        // Directly access the Name field of the Pokemon structure within the encounter.
        fmt.Println("- ", encounterCopy.Pokemon.Name)
        // Additional asynchronous operations can be added here.
    }()
}

// Wait for all goroutines to complete before continuing.
wg.Wait()

}

func CommandCatch(name string, cache *pokecache.Cache, pokemon *Pokemon, pokemonURl string) error{
	url := pokemonURl + name
	if data, found := cache.Get(url); found {
		// Data is found in the cache, unmarshal and use it
		if err := json.Unmarshal(data, pokemon); err != nil {
			fmt.Println("Error decoding JSON from cache:", err)
			return err
		}
    } else {
		// Data not found in cache, proceed with the HTTP request
		resp , err := http.Get(url)
		if err != nil{
			fmt.Println("Error fetching data:", err)
			return nil
		}
		defer resp.Body.Close()

		// Use io.ReadAll here
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return nil
		}
		//unmarshal the JSON data into the struct
		err = json.Unmarshal(body, &pokemon)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
			return nil
		}
		cache.Set(url, body, 30*time.Minute)
	}

	if _, found := cache.Get("caught:" + name); found {
		fmt.Printf("%s was already caught! Please try with another Pokemon!\n", name)
	} else {
		chance := rand.Float64() // Generates a number between 0.0 and 1.0
		catchRate := 0.5 - float64(pokemon.BaseExperience)/1000.0
		if chance > catchRate {
            fmt.Printf("You caught %s!\n", name)
			data, _ := json.Marshal(pokemon)
			// 100 years
			const longTime = 100 * 365 * 24 * time.Hour 
			cache.Set("caught:" + name, data, longTime)
		} else {
            fmt.Printf("%s escaped!\n", name)
        }

	}
	return nil
}

func CommandInspect(name string, cache *pokecache.Cache, pokemon *Pokemon) error{
	
	if data, found := cache.Get("caught:"+name); found {
		// Data is found in the cache, unmarshal and use it
		if err := json.Unmarshal(data, pokemon); err != nil {
			fmt.Println("Error decoding JSON from cache:", err)
			return err
		}
	} else {
		fmt.Printf("You have not caught %s. Please catch it!!!\n", name)
		return nil
	}

	// Start printing the formatted Pokémon details
    fmt.Printf("Name: %s\n", pokemon.Name)
    fmt.Printf("Height: %d\n", pokemon.Height)
    fmt.Printf("Weight: %d\n", pokemon.Weight)
    fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
        fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
    }
	fmt.Println("Types:")
    for _, t := range pokemon.Types {
        fmt.Printf("  - %s\n", t.Type.Name)
    }
	return nil
}

func CommandPokedex(cache *pokecache.Cache) error {
    fmt.Println("Your Pokedex:")
    
    cache.Iterate(func(key string, value []byte) bool {
        // Assuming caught Pokémon are identified with a "caught:" prefix in their keys
        if strings.HasPrefix(key, "caught:") {
            pokemonName := strings.TrimPrefix(key, "caught:")
            fmt.Println(" -", pokemonName)
        }
        return true // Continue iteration
    })
    
    return nil
}

// RegisterExitCallback allows the main package to register a callback function for exit.
func RegisterExitCallback(callback func()) {
    onExitCallback = callback
}