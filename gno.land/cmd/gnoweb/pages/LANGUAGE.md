# Gno.land Website Content - Gnolang Language

## About the Gnolang (Gno) Language

[Gnolang](https://github.com/gnolang/gno/blob/master/LICENSE.md) (Gno) is an interpretation of the widely-used Golang (Go) programming language for blockchain created by Cosmos co-founder Jae Kwon in 2022. Gno is ~99% identical to Go, so Go programmers can start coding in Gno right away, with a minimal learning curve. For example, Gno comes with blockchain-specific standard libraries, but any code that doesn’t use blockchain-specific logic can run in Go with minimal processing. Libraries that don’t make sense in the blockchain context, such as network or operating-system access, are not available in Gno. Otherwise, Gno loads and uses many standard libraries that power Go, so most of the parsing of the source code is the same.

Under the hood, the Gno code is parsed into an abstract syntax tree (AST) and the AST itself is used in the interpreter, rather than byte code as in many virtual machines such as Java, Python, or WASM. This makes even the Gno VM accessible to any Go programmer. The novel design of the GnoVM interpreter allows Gno to freeze and resume the program by persisting and loading the entire memory state. This allows (smart contract) programs to be succinct, as the programmer doesn’t have to serialize and deserialize objects to persist them into a database (unlike programming applications with the Cosmos SDK).

## How Gno Differs from Go

The composable nature of Go/Gno allows for type-checked interactions between contracts, making Gno.land safer and more powerful, as well as operationally cheaper and faster. Smart contracts on Gno.land are light, simple, more focused, and easily interoperable—a network of interconnected contracts rather than siloed monoliths that limit interactions with other contracts.

## Gno Inherits Go’s Built-in Security Features

Go supports secure programming through exported/non-exported fields, enabling a “least-authority” design. It is easy to create objects and APIs that expose only what should be accessible to callers while hiding what should not be simply by the capitalization of letters, thus allowing a succinct representation of secure logic that can be called by multiple users.

Another major advantage of Go is that the language comes with an ecosystem of great tooling, like the compiler and third-party tools that statically analyze code. Gno inherits these advantages from Go directly to create a smart contract programming language that is safe and helps developers to write secure code relying on the compiler, parser, and interpreter to give warning alerts for common mistakes.

## Gno Is Essential for the Wider Adoption of Web3

Other smart contract programming languages like Solidity are designed for one purpose only (writing smart contracts) and are bound by the limitations of the EVM. In addition, developers have to learn several languages if they want to understand the whole stack or work across different ecosystems. On Gno.land, every part of the stack is written in Gno so it’s easy to understand the entire system just by studying a relatively small code base.

Using Gno, developers can rapidly accelerate application development and adopt a modular structure by reusing and reassembling existing modules without building from scratch. They can embed one structure inside another in an intuitive way while preserving localism, and the language specification is simple, successfully balancing practicality and minimalism.

The Go language is so well designed that the Gno smart contract system will become the new gold standard for smart contract development and other blockchain applications. As a programming language that is universally adopted, secure, composable, and complete, Gno is essential for the broader adoption of web3 and its sustainable growth.
