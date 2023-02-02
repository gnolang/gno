package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func twitterFetchTips() string {
	if opts.since != "" {
		opts.twitterSearchTweetsUrl += "&start_time=" + opts.since
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

type TweetSearch struct {
	Data     []Tweet `json:"data"`
	Meta     Info    `json:"meta"`
	Includes Include `json:"includes"`
}

type Tweet struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	AuthorId  string `json:"author_id"`
}

type Info struct {
	NewestId    string `json:"newest_id"`
	OldestId    string `json:"oldest_id"`
	ResultCount int    `json:"result_count"`
}

type Include struct {
	Users []User `json:"users"`
}

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}
