package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"city-tags-api/internal/api_errors"
	"city-tags-api/internal/database"
)

type Api struct {
	db database.Service
}

func NewServer(port int) *http.Server {
	api := &Api{
		db: database.New(),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      api.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
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

	rows, err := api.db.Query(fmt.Sprintf("select * from city_tags.cities where city_id = %d", cityId))
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
