package main

import (
	"bytes"
	"encoding/json"
	"os"
)

func jsonFormat(data string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(data), "", "\t")
	if err != nil {
		return data
	}
	return out.String()
}

func writeOutputFiles(outputs map[string]string) error {
	if _, err := os.Stat(opts.outputPath); os.IsNotExist(err) {
		err = os.MkdirAll(opts.outputPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	for name, data := range outputs {
		if data == "" {
			continue
		}
		if opts.format == "json" {
			data = jsonFormat(data)
		}
		err := os.WriteFile(opts.outputPath+name+".json", []byte(data), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
