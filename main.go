package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		do_cliCommand(cleanInput(scanner.Text())[0])
		//clean_text := cleanInput(scanner.Text())
		//command := clean_text[0]
		//fmt.Printf("Your command was: %s\n", clean_text[0])
	}
}
