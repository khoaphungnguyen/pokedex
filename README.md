# Pokedex CLI

The Pokedex CLI is a command-line interface application that simulates the functionality of a Pokédex from the Pokémon series, allowing users to explore, catch, and inspect Pokémon with details such as names, types, and stats.

## Features

- **Explore**: Discover Pokémon in various areas with `explore [area-name]`.
- **Catch**: Add Pokémon to your Pokédex using `catch [pokemon-name]`.
- **Inspect**: View detailed information of caught Pokémon with `inspect [pokemon-name]`.
- **Pokedex**: List all Pokémon you have caught with `pokedex`.
- **Help**: Get a list of available commands with `help`.
- **Exit**: Save progress and exit the application gracefully with `exit`.

## Usage

1. **Launch**: Run the application in your terminal. You will see the `Pokedex>` prompt.
2. **Explore Areas**: Type `explore kanto-forest` to see Pokémon in "kanto-forest".
3. **Catch Pokémon**: Use `catch pikachu` to attempt catching Pikachu.
4. **Inspect Pokémon**: After catching, inspect them with `inspect bulbasaur`.
5. **View Your Pokedex**: Enter `pokedex` to see your collection.
6. **Exit**: Use `exit` to save progress and exit the application.

## Technical Details

- **Caching**: Utilizes a custom caching mechanism with mutexes to speed up data retrieval and reduce API calls. This ensures faster response times for previously fetched data and enhances the user experience by providing immediate access to Pokémon details.
- **Concurrency**: The cache is designed with concurrency in mind, using mutexes to ensure thread-safe access. This allows multiple read operations to occur simultaneously without blocking, leveraging Go's concurrency model to efficiently handle requests.
- **Persistence**: Progress is saved in a `pokedex.json` file in the application's running directory. The application ensures data is persisted between sessions, allowing users to pick up where they left off.
- **Signal Handling**: Implements graceful shutdown and signal handling to ensure user progress is saved, even if the application is interrupted (e.g., Ctrl+C).

## Requirements

- Go 1.22 or later
- Internet connection (for initial data fetch from PokeAPI)

## Installation

1. Clone the repository.
2. Navigate to the project directory and run `go build`.
3. Start the application with `./pokedex`
