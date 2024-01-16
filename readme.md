### Lox Scripting Language Interpreter

A version of the Tree-Walk Interpreter written in Go.
Based on the book https://craftinginterpreters.com originally coded in Java.

#### Additions to Lox

* Lexer
  * C-style multiline comments `/* ... */`
  * `break` keyword
* Parser
  * `break` statement
* Resolver
  * ParseError: unused local variable
* Interpreter
  * handle `break` statement in `for` and `while` loops
  * handle return statement via state instead of with exception handling (~4 times faster)
