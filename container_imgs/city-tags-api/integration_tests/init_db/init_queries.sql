\COPY city_tags.cities (city_id, city_name, continent, country_3_code) FROM './integration_tests/init_db/test_data/cities_table.csv' DELIMITER ',' CSV HEADER;

\COPY city_tags.city_tags (city_id, cloud_coverage_tag, humidity_tag, temp_tag, precipitation_tag, air_quality_tag, daylight_hours_tag, city_size_tag) FROM './integration_tests/init_db/test_data/city_tags.csv' DELIMITER ',' CSV HEADER;
