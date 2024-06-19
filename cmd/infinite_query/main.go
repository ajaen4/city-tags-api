package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	apiEndpoint := "http://city-tags-api-lb-1927280828.eu-west-1.elb.amazonaws.com/v0/cities?limit=10&offset=200"
	interval := 5 * time.Second

	for {
		err := makeRequest(apiEndpoint)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
		}

		time.Sleep(interval)
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
