-- +goose Up
-- +goose StatementBegin
CREATE TABLE city_tags.cities (
    city_id INT PRIMARY KEY,
    city_name VARCHAR(50) NOT NULL,
    continent VARCHAR(50) NOT NULL,
    country_3_code VARCHAR(3) NOT NULL
);

CREATE TABLE city_tags.city_tags (
    city_id INT PRIMARY KEY,
    cloud_coverage_tag	VARCHAR(30) NOT NULL,
    humidity_tag VARCHAR(30) NOT NULL,
    temp_tag VARCHAR(30) NOT NULL,
    precipitation_tag VARCHAR(30) NOT NULL,
    air_quality_tag VARCHAR(30) NOT NULL,
    daylight_hours_tag VARCHAR(30) NOT NULL,
    city_size_tag VARCHAR(30) NOT NULL,
    FOREIGN KEY (city_id) REFERENCES city_tags.cities(city_id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE city_tags.cities;
DROP TABLE city_tags.city_tags;
-- +goose StatementEnd
