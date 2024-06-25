package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyPathError(t *testing.T) {
	t.Parallel()

	cfg := &cfg{
		docsPath: "",
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFn()

	assert.ErrorIs(t, execLint(cfg, ctx), errEmptyPath)
}

func TestExtractLinks(t *testing.T) {
	t.Parallel()

	// Create mock file content with random links
	mockFileContent := `# Lorem Ipsum
Lorem ipsum dolor sit amet, 
[consectetur](https://example.org)
adipiscing elit. Vivamus lacinia odio
vitae [vestibulum vestibulum](http://localhost:3000).
Cras [vel ex](http://192.168.1.1) et
turpis egestas luctus. Nullam
[eleifend](https://www.wikipedia.org)
nulla ac [blandit tempus](https://gitlab.org). 
## Valid Links Here are some valid links:
- [Mozilla](https://mozilla.org) 
- [Valid URL](https://valid-url.net) 
- [Another Valid URL](https://another-valid-url.info) 
- [Valid Link](https://valid-link.edu)
`

	// Expected URLs
	expectedUrls := []string{
		"https://example.org",
		"http://192.168.1.1",
		"https://www.wikipedia.org",
		"https://gitlab.org",
		"https://mozilla.org",
		"https://valid-url.net",
		"https://another-valid-url.info",
		"https://valid-link.edu",
	}

	// Extract URLs from each file in the sourceDir
	extractedUrls := extractUrls([]byte(mockFileContent))

	if len(expectedUrls) != len(extractedUrls) {
		t.Fatal("did not extract correct amount of URLs")
	}

	sort.Strings(extractedUrls)
	sort.Strings(expectedUrls)

	for i, u := range expectedUrls {
		require.Equal(t, u, extractedUrls[i])
	}
}

func TestExtractJSX(t *testing.T) {
	t.Parallel()

	// Create mock file content with random JSX tags
	mockFileContent := `
#### Usage

### getFunctionSignatures

Fetches public facing function signatures

#### Parameters

Returns **Promise<FunctionSignature[]>**

# test text from gnodev.md <node-rpc-listener>

#### Usage
### evaluateExpression

Evaluates any expression in readonly mode and returns the results

#### Parameters

Returns **Promise<string>**
`

	// Expected JSX tags
	expectedTags := []string{
		"<FunctionSignature[]>",
		"<string>",
		"<node-rpc-listener>",
	}

	// Extract JSX tags from the mock file content
	extractedTags := extractJSX([]byte(mockFileContent))

	if len(expectedTags) != len(extractedTags) {
		t.Fatal("did not extract the correct amount of JSX tags")
	}

	sort.Strings(extractedTags)
	sort.Strings(expectedTags)

	for i, tag := range expectedTags {
		require.Equal(t, tag, extractedTags[i])
	}
}

func TestExtractLocalLinks(t *testing.T) {
	t.Parallel()

	// Create mock file content with random local links
	mockFileContent := `
Here is some text with a link to a local file: [text](../concepts/file1.md)
Here is another local link: [another](./path/to/file1.md)
Here is another local link: [another](./path/to/file2.md#header-1-2)
And a link to an external website: [example](https://example.com)
And a websocket link: [websocket](ws://example.com/socket)
Here's an embedmd link: [embedmd]:# (../assets/how-to-guides/simple-library/tapas.gno go)
Here's an embedmd link: [embedmd]:# (../assets/myfile.sol go)
Here's an embedmd link: [embedmd]:# (../assets/myfi()le.gno c)
Here's an embedmd link: [embedmd]:# (../assets/)myfi(le.gno c)
Here's another link: [embedmd]:# (../folder/myfile.gno c
`

	// Expected local links
	expectedLinks := []string{
		"../concepts/file1.md",
		"./path/to/file1.md",
		"./path/to/file2.md",
		"../assets/how-to-guides/simple-library/tapas.gno",
		"../assets/myfile.sol",
		"../assets/myfi()le.gno",
		"../assets/)myfi(le.gno",
		"../folder/myfile.gno",
	}

	// Extract local links tags from the mock file content
	extractedLinks := extractLocalLinks([]byte(mockFileContent))

	if len(expectedLinks) != len(extractedLinks) {
		t.Fatal("did not extract the correct amount of local links")
	}

	sort.Strings(extractedLinks)
	sort.Strings(expectedLinks)

	for i, tag := range expectedLinks {
		require.Equal(t, tag, extractedLinks[i])
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

func TestFlow(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp(".", "test")
	require.NoError(t, err)
	t.Cleanup(removeDir(t, tempDir))

	contents := `This is a [broken Wikipedia link](https://www.wikipedia.org/non-existent-page).
Here's an embedmd link that links to a non-existing file: [embedmd]:# (../assets/myfile.sol go)
and here is some JSX tags <string\> <random-unescaped-text-tag>
and a local link will be added in further code.
`

	filePath := filepath.Join(tempDir, "examplefile.md")

	err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	require.NoError(t, err)

	f, err := os.Create(filePath)
	require.NoError(t, err)

	_, err = f.WriteString(contents)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	filePathToLink := filepath.Join(tempDir, "fileToLinkTo.md")

	contents += fmt.Sprintf(" [This](%s) is a valid local link.", filePathToLink)

	err = execLint(&cfg{
		docsPath: tempDir,
	},
		context.Background(),
	)

	assert.ErrorIs(t, err, errFoundLintItems)
}

func removeDir(t *testing.T, dirPath string) func() {
	return func() {
		require.NoError(t, os.RemoveAll(dirPath))
	}
}
