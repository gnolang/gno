package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func twitterFetchTips() string {
	if opts.from != "" {
		opts.twitterSearchTweetsUrl += "&start_time=" + opts.from
	}

	var bearer = "Bearer " + opts.twitterToken
	req, err := http.NewRequest("GET", opts.twitterSearchTweetsUrl, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return ""
	}
	req.Header.Add("Authorization", bearer)
	resp, err := opts.httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return ""
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return ""
	}
	return string(body)
}
