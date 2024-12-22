package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

func main() {
	filePath := flag.String("file", "gnolive.gno", "path to the Go file to monitor")
	// XXX: query VS tx mode
	flag.Parse()

	// monitor file changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("File modified:", event.Name)
					processFile(*filePath)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()
	err = watcher.Add(*filePath)
	if err != nil {
		log.Fatal(err)
	}

	processFile(*filePath)

	// XXX: os.Exec("$EDITOR", scriptFile)

	// Keep the program running
	<-make(chan struct{})
}

// processFile reads the file, parses it with the Go AST, removes comments, and prints the cleaned file to stdout
func processFile(filePath string) {
	// Read the file
	src, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new file set
	fset := token.NewFileSet()

	// Parse the file
	node, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// Sanitize the file
	// XXX: remove useless imports
	// XXX: run gnolint?
	node.Comments = nil // Remove comments
	// XXX: verify syntax

	// Print the cleaned file to stdout
	err = printer.Fprint(os.Stdout, fset, node)
	if err != nil {
		log.Fatal(err)
	}
	// XXX: os.Exec("gnokey maketx run -file=", src)

}
