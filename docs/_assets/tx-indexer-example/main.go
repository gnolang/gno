package main

import (
	"fmt"
	"os"
)

func main() {
	mode := ""
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "1", "process":
		fmt.Println("Running processing example")
		RunProcessExample()
	case "2", "api":
		fmt.Println("Running API example (Ctrl+C to quit)")
		RunApiExample()
	case "3", "ws", "websocket":
		fmt.Println("Running websocket example (Ctrl+C to quit)")
		RunWebSocketExample()
	default:
		fmt.Println("Usage: go run . <mode>")
		fmt.Println("Modes:")
		fmt.Println("  1 | process    Parse & display transactions")
		fmt.Println("  2 | api        Run HTTP /stats server")
		fmt.Println("  3 | ws         Live subscription over WebSocket")
	}
}
