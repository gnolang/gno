package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

func lintJSX(filesToAnalyze []string, ctx context.Context) error {
	//file, err := os.Open(filesToAnalyze)
	//if err != nil {
	//	return err
	//}
	//
	//cleanup := func() error {
	//	if closeErr := file.Close(); closeErr != nil {
	//		return fmt.Errorf("unable to gracefully close file, %w", closeErr)
	//	}
	//	return nil
	//}
	//
	//return cleanup()
	return nil
}

func findJSXTags(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	cleanup := func() error {
		if closeErr := file.Close(); closeErr != nil {
			return fmt.Errorf("unable to gracefully close file, %w", closeErr)
		}
		return nil
	}

	scanner := bufio.NewScanner(file)
	jsxTags := make(map[string]string)

	// Scan file line by line
	for scanner.Scan() {
	}

	return jsxTags, cleanup()
}
