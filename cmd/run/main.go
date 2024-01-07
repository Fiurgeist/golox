package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fiurgeist/golox/internal/interpreter"
	"github.com/fiurgeist/golox/internal/lexer"
	"github.com/fiurgeist/golox/internal/parser"
	"github.com/fiurgeist/golox/internal/reporter"
	"github.com/fiurgeist/golox/internal/resolver"
)

// https://man.freebsd.org/cgi/man.cgi?query=sysexits
const (
	EX_OK       = 0
	EX_USAGE    = 64
	EX_DATAERR  = 65
	EX_SOFTWARE = 70
)

const DEBUG = false

var PERF = false
var environment = interpreter.NewEnvironment()

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

	start := time.Now().UnixNano()
	tokens, errLex := lexer.ScanTokens()
	printPerf("Lexing", start)

	if DEBUG {
		for _, token := range tokens {
			fmt.Println(token.String())
		}
	}

	parser := parser.NewParser(tokens, reporter)

	start = time.Now().UnixNano()
	statements, errParse := parser.Parse()
	printPerf("Parsing", start)

	if errLex != nil || errParse != nil || reporter.HadError {
		return EX_DATAERR
	}

	interpreter := interpreter.NewInterpreter(environment, reporter)
	resolver := resolver.NewResolver(interpreter, reporter)

	start = time.Now().UnixNano()
	resolver.Resolve(statements)
	printPerf("Resolveing", start)

	if reporter.HadError {
		return EX_DATAERR
	}

	start = time.Now().UnixNano()
	err := interpreter.Interpret(statements)
	printPerf("Interpreting", start)

	if err != nil {
		return EX_SOFTWARE
	}

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
	PERF = true

	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	code := run(file)
	if code != EX_OK {
		os.Exit(code)
	}
}

func printPerf(operation string, start int64) {
	if !PERF {
		return
	}

	fmt.Printf("%s took %fms\n", operation, float64(time.Now().UnixNano()-start)/1000000.0)
}
