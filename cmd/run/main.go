package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/fiurgeist/golox/internal/ast"
	"github.com/fiurgeist/golox/internal/interpreter"
	"github.com/fiurgeist/golox/internal/lexer"
	"github.com/fiurgeist/golox/internal/parser"
	"github.com/fiurgeist/golox/internal/reporter"
)

// https://man.freebsd.org/cgi/man.cgi?query=sysexits
const (
	EX_OK       = 0
	EX_USAGE    = 64
	EX_DATAERR  = 65
	EX_SOFTWARE = 70
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

func run(script []byte) int {
	reporter := &reporter.ConsoleReporter{}
	lexer := lexer.NewLexer(script, reporter)
	tokens, errLex := lexer.ScanTokens()

	for _, token := range tokens {
		fmt.Println(token.String())
	}

	parser := parser.NewParser(tokens, reporter)
	expr, errParse := parser.Parse()

	if errLex != nil || errParse != nil {
		return EX_DATAERR
	}

	err := interpreter.Interpret(expr, reporter)
	if err != nil {
		return EX_SOFTWARE
	}

	fmt.Println(ast.Print(expr))
	return EX_OK
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
	code := run(file)
	if code != EX_OK {
		os.Exit(code)
	}
}
