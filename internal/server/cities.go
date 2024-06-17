package server

import (
	"city-tags-api/internal/api_errors"
	"net/http"
	"strconv"
)

type getCitiesResp struct {
	Cities []getCityResp `json:"cities"`
	Offset int           `json:"offset"`
}

// @Summary		Get cities
// @Description	Get cities with pagination
// @Accept		json
// @Produce		json
// @Param       offset  query int	false	"Offset for pagination"
// @Param       limit   query int	false	"Limit for pagination"
// @Success		200 	{object} 	getCityResp
// @Failure     500 	{object} 	api_errors.ClientErr
// @Router		/v0/cities/ [get]
func (api *Api) getCities(w http.ResponseWriter, r *http.Request) error {
	offsetParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetParam)
	if offsetParam == "" || err != nil {
		offset = 0
	}

	limitParam := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitParam)
	if limitParam == "" || err != nil {
		limit = 100
	}

	rows, err := api.db.Query("select * from city_tags.cities LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return err
	}

	var citiesData = getCitiesResp{
		Cities: []getCityResp{},
		Offset: offset + limit,
	}
	for rows.Next() {
		cityData := getCityResp{}
		err = rows.Scan(&cityData.CityId, &cityData.CityName, &cityData.Continent, &cityData.Country3Code)
		if err != nil {
			return err
		}
		citiesData.Cities = append(citiesData.Cities, cityData)
	}

	respondWithJSON(w, http.StatusOK, citiesData)
	return nil
}

type getCityResp struct {
	CityId       int    `json:"city_id"`
	CityName     string `json:"city_name"`
	Continent    string `json:"continent"`
	Country3Code string `json:"country_3_code"`
}

// @Summary		Get city by city id
// @Description	Get city information by providing a specific city id
// @Accept			json
// @Produce		json
// @Success		200 {object} getCityResp
// @Failure      500  {object} api_errors.ClientErr
// @Router			/v0/cities/{cityId} [get]
func (api *Api) getCity(w http.ResponseWriter, r *http.Request) error {
	cityIdParam := r.PathValue("cityId")
	cityId, err := strconv.Atoi(cityIdParam)
	if cityIdParam == "" || err != nil {
		return &api_errors.ClientErr{
			HttpCode: http.StatusBadRequest,
			Message:  "cityId parameter not present or invalid",
		}
	}

	rows, err := api.db.Query("select * from city_tags.cities where city_id = $1", cityId)
	if err != nil {
		return err
	}

	var cityData getCityResp
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
