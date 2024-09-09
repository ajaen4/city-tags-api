package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	endpoint := os.Getenv("ENDPOINT")
	if endpoint == "" {
		log.Fatal("ENDPOINT environment variable is not set")
	}

	token := os.Getenv("JWT")
	if token == "" {
		log.Fatal("JWT environment variable is not set")
	}

	err := makeRequest(endpoint, token)
	if err != nil {
		log.Fatalf("Error making request: %v\n", err)
	}
}

func makeRequest(url, token string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	log.Println("Request successful")
	log.Println("Response body:", string(body))

	return nil
}
