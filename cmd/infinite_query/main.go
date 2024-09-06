package main

import (
	"fmt"
	"net/http"
)

func main() {
	apiEndpoint := "https://dev.city-tags-api.sityex.com/v0/cities?limit=10&offset=200"

	for {
		err := makeRequest(apiEndpoint)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
		}
	}
}

func makeRequest(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Println("Request successful")

	return nil
}
