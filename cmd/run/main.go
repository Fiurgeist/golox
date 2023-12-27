package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/fiurgeist/golox/internal/lexer"
)

const (
	EX_USAGE   = 64
	EX_DATAERR = 65
)

func main() {
	if len(os.Args) > 2 {
		fmt.Print("Usage: golox [script]\n")
		os.Exit(EX_USAGE)
	}

	if len(os.Args) == 1 {
		runPrompt()
		return
	}

	path := os.Args[1]
	runFile(path)
}

func run(script string) {
	lexer := lexer.NewLexer(script)
	tokens, err := lexer.ScanTokens()
	if err != nil {
		error(0, err.Error())
		os.Exit(EX_DATAERR)
	}

	for _, token := range tokens {
		fmt.Println(token.String())
	}
}

func runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Lox REPL")

	for {
		fmt.Print("> ")

		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}

		line := scanner.Text()
		if line == "" {
			// the typical Ctrl-D to exit will cause an empty line -> break
			fmt.Println("")
			break
		}

		run(line)
	}
}

func runFile(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	run(string(file))
}

func error(line int32, message string) {
	report(line, "", message)
}

func report(line int32, where, message string) {
	fmt.Printf("[line %d] Error%s: %s", line, where, message)
}
