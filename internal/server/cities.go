package server

import (
	"city-tags-api/internal/api_errors"
	"fmt"
	"net/http"
	"strconv"
)

type GetCityResp struct {
	CityId       int    `json:"city_id"`
	CityName     string `json:"city_name"`
	Continent    string `json:"continent"`
	Country3Code string `json:"country_3_code"`
}

type GetCitiesResp struct {
	Cities []GetCityResp `json:"cities"`
	Offset int           `json:"offset"`
}

type getCitiesReq struct {
	offset int
	limit  int
}

func (getCitR *getCitiesReq) validate(r *http.Request) error {
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
// @Success		200 	{object} 	getCityResp
// @Failure     500 	{object} 	api_errors.ClientErr
// @Router		/v0/cities [get]
func (api *Api) getCities(w http.ResponseWriter, r *http.Request) error {
	citiesReq := &getCitiesReq{}
	err := citiesReq.validate(r)
	if err != nil {
		return err
	}

	rows, err := api.db.Query("select * from city_tags.cities LIMIT $1 OFFSET $2", citiesReq.limit, citiesReq.offset)
	if err != nil {
		return err
	}

	var citiesData = GetCitiesResp{
		Cities: []GetCityResp{},
		Offset: citiesReq.offset + citiesReq.limit,
	}
	for rows.Next() {
		cityData := GetCityResp{}
		err = rows.Scan(&cityData.CityId, &cityData.CityName, &cityData.Continent, &cityData.Country3Code)
		if err != nil {
			return err
		}
		citiesData.Cities = append(citiesData.Cities, cityData)
	}

	respondWithJSON(w, http.StatusOK, citiesData)
	return nil
}

type getCityReq struct {
	cityId int
}

func (getCitR *getCityReq) validate(r *http.Request) error {
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
// @Success		200 {object} getCityResp
// @Failure      500  {object} api_errors.ClientErr
// @Router			/v0/cities/{cityId} [get]
func (api *Api) getCity(w http.ResponseWriter, r *http.Request) error {
	getCityReq := getCityReq{}
	err := getCityReq.validate(r)
	if err != nil {
		return err
	}

	rows, err := api.db.Query("select * from city_tags.cities where city_id = $1", getCityReq.cityId)
	if err != nil {
		return err
	}

	var cityData GetCityResp
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
