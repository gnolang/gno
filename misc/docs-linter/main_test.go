package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestEmptyPathError(t *testing.T) {
	t.Parallel()

	cfg := &cfg{
		docsPath: "",
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFn()

	err := execLint(cfg, ctx)
	assert.ErrorIs(t, err, errEmptyPath)
}

func TestExtractLinks(t *testing.T) {
	t.Parallel()

	// Generate temporary source dir
	sourceDir, err := os.MkdirTemp(".", "sourceDir")
	require.NoError(t, err)
	t.Cleanup(removeDir(t, sourceDir))

	// Create mock files with random links
	mockFiles := map[string]string{
		"file1.md": "This is a test file with a link: https://example.com.\nAnother link: http://example.org.",
		"file2.md": "Markdown content with a link: https://example.com/page.",
		"file3.md": "Links in a list:\n- https://example.com/item1\n- https://example.org/item2",
	}

	for fileName, content := range mockFiles {
		filePath := filepath.Join(sourceDir, fileName)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Expected URLs and their corresponding files
	expectedUrls := map[string]string{
		"https://example.com":       filepath.Join(sourceDir, "file1.md"),
		"http://example.org":        filepath.Join(sourceDir, "file1.md"),
		"https://example.com/page":  filepath.Join(sourceDir, "file2.md"),
		"https://example.com/item1": filepath.Join(sourceDir, "file3.md"),
		"https://example.org/item2": filepath.Join(sourceDir, "file3.md"),
	}

	// Extract URLs from each file in the sourceDir
	for fileName := range mockFiles {
		filePath := filepath.Join(sourceDir, fileName)
		extractedUrls, err := extractUrls(filePath)
		require.NoError(t, err)

		// Verify that the extracted URLs match the expected URLs
		for url, expectedFile := range expectedUrls {
			if expectedFile == filePath {
				require.Equal(t, expectedFile, extractedUrls[url], "URL: %s not correctly mapped to file: %s", url, expectedFile)
			}
		}
	}
}

func TestFindFilePaths(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp(".", "test")
	require.NoError(t, err)
	t.Cleanup(removeDir(t, tempDir))

	numSourceFiles := 20
	testFiles := make([]string, numSourceFiles)

	for i := 0; i < numSourceFiles; i++ {
		testFiles[i] = "sourceFile" + strconv.Itoa(i) + ".md"
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		require.NoError(t, err)

		_, err = os.Create(filePath)
		require.NoError(t, err)
	}

	results, err := findFilePaths(tempDir)
	require.NoError(t, err)

	expectedResults := make([]string, 0, len(testFiles))

	for _, testFile := range testFiles {
		expectedResults = append(expectedResults, filepath.Join(tempDir, testFile))
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})

	sort.Slice(expectedResults, func(i, j int) bool {
		return expectedResults[i] < expectedResults[j]
	})

	require.Equal(t, len(results), len(expectedResults))

	for i, result := range results {
		if result != expectedResults[i] {
			require.Equal(t, result, expectedResults[i])
		}
	}
}

func removeDir(t *testing.T, dirPath string) func() {
	return func() {
		err := os.RemoveAll(dirPath)
		require.NoError(t, err)
	}
}
