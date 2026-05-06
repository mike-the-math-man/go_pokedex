# GoDex

## Description
GoDex is a command-line Pokedex application built in Go. It provides an interactive REPL (Read-Eval-Print Loop) interface where users can explore Pokemon data, including details about individual Pokemon, their types, abilities, and more. The application fetches real-time data from the PokeAPI to ensure up-to-date information.

## Motivation
This project was created as a portfolio piece to demonstrate proficiency in Go programming, particularly focusing on:
- Building CLI applications with interactive user interfaces
- Implementing caching mechanisms for API responses
- Handling HTTP requests and JSON parsing
- Structuring clean, maintainable Go code
- Working with external APIs

## Quick Start
1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/go_pokedex.git
   cd go_pokedex
   ```

2. **Install Go (if not already installed):**
   - Download from [golang.org](https://golang.org/dl/)

3. **Build the application:**
   ```bash
   go build -o pokedex main.go
   ```

4. **Run the application:**
   ```bash
   ./pokedex
   ```

## Usage
Once running, you'll enter an interactive REPL. Here are some available commands:

- `help` - Display available commands
- `exit` - Exit the application
- `map` - Display the next 20 location areas
- `mapb` - Display the previous 20 location areas
- `explore <location>` - Explore a specific location area
- `catch <pokemon>` - Attempt to catch a Pokemon
- `inspect <pokemon>` - View details of a caught Pokemon
- `pokedex` - List all caught Pokemon

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request