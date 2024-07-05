package server

import (
	"city-tags-api/internal/server"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func TestGetCity(t *testing.T) {
	tests := []struct {
		name     string
		cityId   string
		expected server.GetCityResp
		isError  bool
	}{
		{
			"City found",
			"3838859",
			server.GetCityResp{
				CityId: 3838859, CityName: "Río Gallegos", Continent: "South America", Country3Code: "ARG",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := fmt.Sprintf("http://localhost:8080/v0/cities/%s", tt.cityId)
			resp, err := http.Get(endpoint)
			if !tt.isError && err != nil {
				t.Fatal(err)
			} else if tt.isError && err == nil {
				t.Errorf("Expected error but none was received")
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			fmtResp := &server.GetCityResp{}
			err = json.Unmarshal(body, fmtResp)
			if err != nil {
				t.Fatal(err)
			}

			if *fmtResp != tt.expected {
				t.Errorf("%s returned %v want %v", endpoint, *fmtResp, tt.expected)
			}
		})
	}
}
