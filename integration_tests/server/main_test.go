package server

import (
	"city-tags-api/internal/api_errors"
	"city-tags-api/internal/server"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var endpoint string
var testJWT string

func TestMain(m *testing.M) {
	serverHost := os.Getenv("SERVER_HOST")
	serverPort := os.Getenv("SERVER_PORT")
	endpoint = fmt.Sprintf("http://%s:%s", serverHost, serverPort)

	testJWT = os.Getenv("TEST_JWT")

	code := m.Run()
	os.Exit(code)
}

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
			url := fmt.Sprintf("%s/v0/cities/%s", endpoint, tt.cityId)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", testJWT))

			client := &http.Client{}
			resp, err := client.Do(req)
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
			fmt.Sprintf("%s/v0/cities%s", endpoint, "?offset=0"),
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
			fmt.Sprintf("%s/v0/cities%s", endpoint, "?offset=1"),
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
			fmt.Sprintf("%s/v0/cities%s", endpoint, "?limit=1"),
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
			fmt.Sprintf("%s/v0/cities%s", endpoint, "?offset=1&limit=1"),
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
			req, err := http.NewRequest("GET", tt.URL, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", testJWT))

			client := &http.Client{}
			resp, err := client.Do(req)
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
					t.Errorf("%s returned city %v want %v", tt.URL, city, tt.expected.Cities[index])
				}
			}

			if fmtResp.Offset != tt.expected.Offset {
				t.Errorf("%s returned offset %d want %d", tt.URL, fmtResp.Offset, tt.expected.Offset)
			}
		})
	}
}
