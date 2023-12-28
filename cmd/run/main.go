package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/fiurgeist/golox/internal/lexer"
	"github.com/fiurgeist/golox/internal/reporter"
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

func run(script []byte) {
	reporter := &reporter.ConsoleReporter{}
	lexer := lexer.NewLexer(script, reporter)
	tokens, err := lexer.ScanTokens()
	if err != nil {
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

		run([]byte(line))
	}
}

func runFile(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	run(file)
}
