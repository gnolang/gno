package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mvdan.cc/xurls/v2"
)

func main() {
	root := "../"
	rxRelaxed := xurls.Relaxed()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".md") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return err
			}

			urls := rxRelaxed.FindAllString(string(content), -1)

			if len(urls) > 0 {
				for _, url := range urls {
					resp, err := client.Get(url)
					if err != nil {
						fmt.Printf("Error retrieving URL %s: %v\n", url, err)
						continue
					}

					defer resp.Body.Close()

					if resp.StatusCode == 404 {
						fmt.Printf("Broken link (404) found in file %s: %s\n", path, url)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking the directory tree:", err)
	}
}
