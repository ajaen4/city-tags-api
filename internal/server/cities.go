package server

import (
	"city-tags-api/internal/api_errors"
	"fmt"
	"net/http"
	"strconv"
)

type CityData struct {
	CityId       int    `json:"city_id"`
	CityName     string `json:"city_name"`
	Continent    string `json:"continent"`
	Country3Code string `json:"country_3_code"`
}

type GetCitiesResp struct {
	Cities []CityData `json:"cities"`
	Offset int        `json:"offset"`
}

type TagsData struct {
	CityId        int    `json:"city_id"`
	CloudCoverage string `json:"cloud_coverage"`
	Humidity      string `json:"humidity"`
	Temp          string `json:"temperature"`
	Precipitation string `json:"precipitation"`
	AirQuality    string `json:"air_quality"`
	DaylightHours string `json:"daylight_hours"`
	CitySize      string `json:"city_size"`
}

type GetTagsResp struct {
	Tags map[string]string `json:"tags"`
}

type GetCitiesReq struct {
	offset int
	limit  int
}

func (getCitR *GetCitiesReq) validate(r *http.Request) error {
	var err error
	var offset int
	clientErr := &api_errors.ClientErr{
		HttpCode: http.StatusBadRequest,
		Message:  "Parameters not present or invalid",
	}

	offsetParam := r.URL.Query().Get("offset")
	if offsetParam == "" {
		offset = 0
	} else {
		offset, err = strconv.Atoi(offsetParam)
	}
	if err != nil {
		clientErr.Errors = map[string]string{
			"offset": "Not present or invalid",
		}
		clientErr.LogMess = err.Error()
	}

	var limit int
	limitParam := r.URL.Query().Get("limit")
	if limitParam == "" {
		limit = 100
	} else {
		limit, err = strconv.Atoi(limitParam)
	}
	if err != nil {
		clientErr.Errors = map[string]string{
			"offset": "Not present or invalid",
		}
		if clientErr.LogMess != "" {
			clientErr.LogMess += fmt.Sprintf(", %s", err.Error())
		} else {
			clientErr.LogMess = err.Error()
		}
	}

	if clientErr.LogMess != "" {
		return clientErr
	}

	getCitR.offset = offset
	getCitR.limit = limit
	return nil
}

// @Summary		Get cities
// @Description	Get cities with pagination
// @Accept		json
// @Produce		json
// @Param       offset  query int	false	"Offset for pagination"
// @Param       limit   query int	false	"Limit for pagination"
// @Success		200 	{object} 	CityData
// @Failure     500 	{object} 	api_errors.ClientErr
// @Router		/v0/cities [get]
func (api *Api) getCities(w http.ResponseWriter, r *http.Request) error {
	citiesReq := &GetCitiesReq{}
	err := citiesReq.validate(r)
	if err != nil {
		return err
	}

	rows, err := api.db.Query("select * from city_tags.cities LIMIT $1 OFFSET $2", citiesReq.limit, citiesReq.offset)
	if err != nil {
		return err
	}

	var citiesData = GetCitiesResp{
		Cities: []CityData{},
	}
	count := 0
	for rows.Next() {
		cityData := CityData{}
		err = rows.Scan(&cityData.CityId, &cityData.CityName, &cityData.Continent, &cityData.Country3Code)
		if err != nil {
			return err
		}
		citiesData.Cities = append(citiesData.Cities, cityData)
		count++
	}
	citiesData.Offset = citiesReq.offset + count

	respondWithJSON(w, http.StatusOK, citiesData)
	return nil
}

type GetCityReq struct {
	cityId int
}

func (getCitR *GetCityReq) validate(r *http.Request) error {
	cityIdParam := r.PathValue("cityId")
	cityId, err := strconv.Atoi(cityIdParam)
	if cityIdParam == "" || err != nil {
		return &api_errors.ClientErr{
			HttpCode: http.StatusBadRequest,
			Message:  "Parameters not present or invalid",
			Errors: map[string]string{
				"cityId": "Not present or invalid",
			},
		}
	}
	getCitR.cityId = cityId
	return nil
}

// @Summary		Get city by city id
// @Description	Get city information by providing a specific city id
// @Accept			json
// @Produce		json
// @Success		200 {object} CityData
// @Failure      500  {object} api_errors.ClientErr
// @Router			/v0/cities/{cityId} [get]
func (api *Api) getCity(w http.ResponseWriter, r *http.Request) error {
	GetCityReq := GetCityReq{}
	err := GetCityReq.validate(r)
	if err != nil {
		return err
	}

	rows, err := api.db.Query("select * from city_tags.cities where city_id = $1", GetCityReq.cityId)
	if err != nil {
		return err
	}

	var cityData CityData
	if rows.Next() {
		err = rows.Scan(&cityData.CityId, &cityData.CityName, &cityData.Continent, &cityData.Country3Code)
		if err != nil {
			return err
		}
	} else {
		return &api_errors.ClientErr{
			HttpCode: http.StatusNotFound,
			Message:  "City not found",
		}
	}

	respondWithJSON(w, http.StatusOK, cityData)
	return nil
}

type GetTagsReq struct {
	cityId int
}

func (getTagsR *GetTagsReq) validate(r *http.Request) error {
	cityIdParam := r.PathValue("cityId")
	cityId, err := strconv.Atoi(cityIdParam)
	if cityIdParam == "" || err != nil {
		return &api_errors.ClientErr{
			HttpCode: http.StatusBadRequest,
			Message:  "Parameters not present or invalid",
			Errors: map[string]string{
				"cityId": "Not present or invalid",
			},
		}
	}
	getTagsR.cityId = cityId
	return nil
}

// @Summary		Get city tags by city id
// @Description	Get tags information by providing a specific city id
// @Accept			json
// @Produce		json
// @Success		200 {object} TagsData
// @Failure      500  {object} api_errors.ClientErr
// @Router			/v0/cities/{cityId}/tags [get]
func (api *Api) getTags(w http.ResponseWriter, r *http.Request) error {
	getTagsReq := GetTagsReq{}
	err := getTagsReq.validate(r)
	if err != nil {
		return err
	}

	rows, err := api.db.Query("select * from city_tags.city_tags where city_id = $1", getTagsReq.cityId)
	if err != nil {
		return err
	}

	var tagsData TagsData
	if rows.Next() {
		err = rows.Scan(
			&tagsData.CityId,
			&tagsData.CloudCoverage,
			&tagsData.Humidity,
			&tagsData.Temp,
			&tagsData.Precipitation,
			&tagsData.AirQuality,
			&tagsData.DaylightHours,
			&tagsData.CitySize,
		)
		if err != nil {
			return err
		}
	} else {
		return &api_errors.ClientErr{
			HttpCode: http.StatusNotFound,
			Message:  "City not found",
		}
	}

	respondWithJSON(w, http.StatusOK, tagsData)
	return nil
}
