package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	"os"
	"strings"
)

func jsonFormat(data string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(data), "", "\t")
	if err != nil {
		return data
	}
	return out.String()
}

func createOutputDir() error {
	if _, err := os.Stat(opts.outputPath); os.IsNotExist(err) {
		err = os.MkdirAll(opts.outputPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeChangelog(data string) error {
	err := createOutputDir()
	if err != nil {
		return err
	}
	if opts.format == "json" {
		data = jsonFormat(data)
	}
	err = os.WriteFile(opts.outputPath+"changelog.json", []byte(data), 0644)
	if err != nil {
		return err
	}
	return nil
}

func writeBacklog(data string) error {
	err := createOutputDir()
	if err != nil {
		return err
	}
	if opts.format == "json" {
		data = jsonFormat(data)
	}
	err = os.WriteFile(opts.outputPath+"backlog.json", []byte(data), 0644)
	if err != nil {
		return err
	}
	return nil
}

func writeCuration(data string) error {
	err := createOutputDir()
	if err != nil {
		return err
	}
	if opts.format == "json" {
		data = jsonFormat(data)
	}
	err = os.WriteFile(opts.outputPath+"curation.json", []byte(data), 0644)
	if err != nil {
		return err
	}
	return nil
}

func writeTips(data string) error {
	err := createOutputDir()
	if err != nil {
		return err
	}

	var table [][]string
	var tweets TweetSearch
	_ = json.Unmarshal([]byte(data), &tweets)
	for _, tweet := range tweets.Data {
		table = append(table, []string{tweet.Id, strings.Replace(tweet.Text, "\n", "", -1), tweet.CreatedAt})
	}

	//Maybe build our own table formatter
	markdownTable, err := markdown.NewTableFormatterBuilder().
		Build("Tweet ID", "Text", "Created at").
		Format(table)
	if err != nil {
		return err
	}

	result := fmt.Sprintf("# Tips\n\nThere is **%d new tweet** about gno since %s\n\n%s", tweets.Meta.ResultCount, opts.since, markdownTable)

	err = os.WriteFile(opts.outputPath+"report.md", []byte(result), 0644)
	if err != nil {
		return err
	}
	return nil
}
