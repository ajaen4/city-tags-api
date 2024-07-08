package server

import (
	"city-tags-api/internal/api_errors"
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
		expected any
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
		{
			"City not found",
			"38388599",
			api_errors.ClientErr{
				HttpCode: http.StatusNotFound,
				Message:  "City not found",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := fmt.Sprintf("http://localhost:8080/v0/cities/%s", tt.cityId)
			resp, err := http.Get(endpoint)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if !tt.isError {
				fmtResp := &server.GetCityResp{}
				err = json.Unmarshal(body, fmtResp)
				if err != nil {
					t.Fatal(err)
				}

				if *fmtResp != tt.expected.(server.GetCityResp) {
					t.Errorf("%s returned %v want %v", endpoint, *fmtResp, tt.expected)
				}
			} else {
				fmtResp := &api_errors.ClientErr{}
				err = json.Unmarshal(body, fmtResp)
				if err != nil {
					t.Fatal(err)
				}

				expectedErr := tt.expected.(api_errors.ClientErr)
				if fmtResp.HttpCode != expectedErr.HttpCode || fmtResp.Message != expectedErr.Message {
					t.Errorf("%s returned %v want %v", endpoint, *fmtResp, tt.expected)
				}
			}
		})
	}
}

func TestGetCities(t *testing.T) {
	tests := []struct {
		name     string
		URL      string
		expected server.GetCitiesResp
	}{
		{
			"Get all cities",
			"http://localhost:8080/v0/cities?offset=0",
			server.GetCitiesResp{
				Cities: []server.GetCityResp{
					{
						CityId:       3838859,
						CityName:     "Río Gallegos",
						Continent:    "South America",
						Country3Code: "ARG",
					},
					{
						CityId:       3430443,
						CityName:     "Necochea",
						Continent:    "South America",
						Country3Code: "ARG",
					},
					{
						CityId:       3430988,
						CityName:     "Luján",
						Continent:    "South America",
						Country3Code: "ARG",
					},
				},
				Offset: 3,
			},
		},
		{
			"Test offset",
			"http://localhost:8080/v0/cities?offset=1",
			server.GetCitiesResp{
				Cities: []server.GetCityResp{
					{
						CityId:       3430443,
						CityName:     "Necochea",
						Continent:    "South America",
						Country3Code: "ARG",
					},
					{
						CityId:       3430988,
						CityName:     "Luján",
						Continent:    "South America",
						Country3Code: "ARG",
					},
				},
				Offset: 3,
			},
		},
		{
			"Test limit",
			"http://localhost:8080/v0/cities?limit=1",
			server.GetCitiesResp{
				Cities: []server.GetCityResp{
					{
						CityId:       3838859,
						CityName:     "Río Gallegos",
						Continent:    "South America",
						Country3Code: "ARG",
					},
				},
				Offset: 1,
			},
		},
		{
			"Test limit and offset",
			"http://localhost:8080/v0/cities?offset=1&limit=1",
			server.GetCitiesResp{
				Cities: []server.GetCityResp{
					{
						CityId:       3430443,
						CityName:     "Necochea",
						Continent:    "South America",
						Country3Code: "ARG",
					},
				},
				Offset: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(tt.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			fmtResp := &server.GetCitiesResp{}
			err = json.Unmarshal(body, fmtResp)
			if err != nil {
				t.Fatal(err)
			}

			for index, city := range fmtResp.Cities {
				if tt.expected.Cities[index] != city {
					t.Errorf("%s returned city %v want %v", tt.URL, tt.expected.Cities[index], tt.expected)
				}
			}

			if fmtResp.Offset != tt.expected.Offset {
				t.Errorf("%s returned offset %d want %d", tt.URL, fmtResp.Offset, tt.expected.Offset)
			}
		})
	}
}
