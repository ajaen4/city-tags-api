package server

import (
	"net/http"
	"testing"
)

func TestGetCitiesReq_validate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected getCitiesReq
		isError  bool
	}{
		{"default values", "/v0/cities", getCitiesReq{limit: 100, offset: 0}, false},
		{"explicit values", "/v0/cities?limit=200&offset=200", getCitiesReq{limit: 200, offset: 200}, false},
		{"only offset", "/v0/cities?offset=200", getCitiesReq{limit: 100, offset: 200}, false},
		{"only limit", "/v0/cities?limit=200", getCitiesReq{limit: 200, offset: 0}, false},
		{"incorrect limit", "/v0/cities?limit=200a", getCitiesReq{}, true},
		{"incorrect offset", "/v0/cities?offset=a200", getCitiesReq{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.input, nil)
			if err != nil {
				t.Fatal(err)
			}

			getCitiesReq := &getCitiesReq{}
			err = getCitiesReq.validate(req)
			if tt.isError && err == nil {
				t.Errorf("Expected error but none was received")
			}
			if *getCitiesReq != tt.expected {
				t.Errorf("getCitiesReq.validate(%s) = %v; want %v", tt.input, *getCitiesReq, tt.expected)
			}
		})
	}
}

func TestGetCityReq_validate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected getCityReq
		isError  bool
	}{
		{"incorrect city id", "incorrectCityId", getCityReq{}, true},
		{"correct city id", "3838859", getCityReq{cityId: 3838859}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v0/cities/", nil)
			req.SetPathValue("cityId", tt.input)
			if err != nil {
				t.Fatal(err)
			}

			getCityReq := &getCityReq{}
			err = getCityReq.validate(req)
			if !tt.isError && err != nil {
				t.Errorf("Didn't expect an error but one was received")
			}
			if tt.isError && err == nil {
				t.Errorf("Expected error but none was received")
			}
			if *getCityReq != tt.expected {
				t.Errorf("getCitiesReq.validate(%s) = %v; want %v", tt.input, *getCityReq, tt.expected)
			}
		})
	}
}
