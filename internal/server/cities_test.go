package server

import (
	"net/http"
	"testing"
)

func TestGetCitiesReq_validate(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected GetCitiesReq
		isError  bool
	}{
		{"default values", "/v0/cities", GetCitiesReq{limit: 100, offset: 0}, false},
		{"explicit values", "/v0/cities?limit=200&offset=200", GetCitiesReq{limit: 200, offset: 200}, false},
		{"only offset", "/v0/cities?offset=200", GetCitiesReq{limit: 100, offset: 200}, false},
		{"only limit", "/v0/cities?limit=200", GetCitiesReq{limit: 200, offset: 0}, false},
		{"incorrect limit", "/v0/cities?limit=200a", GetCitiesReq{}, true},
		{"incorrect offset", "/v0/cities?offset=a200", GetCitiesReq{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.input, nil)
			if err != nil {
				t.Fatal(err)
			}

			GetCitiesReq := &GetCitiesReq{}
			err = GetCitiesReq.validate(req)
			if tt.isError && err == nil {
				t.Errorf("Expected error but none was received")
			}
			if *GetCitiesReq != tt.expected {
				t.Errorf("GetCitiesReq.validate(%s) = %v; want %v", tt.input, *GetCitiesReq, tt.expected)
			}
		})
	}
}

func TestGetCityReq_validate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected GetCityReq
		isError  bool
	}{
		{"incorrect city id", "incorrectCityId", GetCityReq{}, true},
		{"correct city id", "3838859", GetCityReq{cityId: 3838859}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/v0/cities/", nil)
			req.SetPathValue("cityId", tt.input)
			if err != nil {
				t.Fatal(err)
			}

			GetCityReq := &GetCityReq{}
			err = GetCityReq.validate(req)
			if !tt.isError && err != nil {
				t.Errorf("Didn't expect an error but one was received")
			}
			if tt.isError && err == nil {
				t.Errorf("Expected error but none was received")
			}
			if *GetCityReq != tt.expected {
				t.Errorf("GetCitiesReq.validate(%s) = %v; want %v", tt.input, *GetCityReq, tt.expected)
			}
		})
	}
}
