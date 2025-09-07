package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type OMDbResponse struct {
	Title    string `json:"Title"`
	Year     string `json:"Year"`
	Poster   string `json:"Poster"`
	Response string `json:"Response"`
	Error    string `json:"Error"`
}

func fetchMovieInfo(title string, year int) (*OMDbResponse, error) {
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OMDB_API_KEY environment variable not set")
	}

	baseURL := "http://www.omdbapi.com/"
	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("t", title)
	if year > 0 {
		params.Add("y", fmt.Sprintf("%d", year))
	}

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie info: %w", err)
	}
	defer resp.Body.Close()

	var result OMDbResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Response == "False" {
		return nil, fmt.Errorf("movie not found: %s", result.Error)
	}

	return &result, nil
}
