// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "City Tags API Support",
            "email": "a.jaenrev@gmail.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/v0/cities": {
            "get": {
                "description": "Get cities with pagination",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get cities",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Offset for pagination",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Limit for pagination",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/server.CityData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api_errors.ClientErr"
                        }
                    }
                }
            }
        },
        "/v0/cities/{cityId}": {
            "get": {
                "description": "Get city information by providing a specific city id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get city by city id",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/server.CityData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api_errors.ClientErr"
                        }
                    }
                }
            }
        },
        "/v0/cities/{cityId}/tags": {
            "get": {
                "description": "Get tags information by providing a specific city id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Get city tags by city id",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/server.TagsData"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/api_errors.ClientErr"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api_errors.ClientErr": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer"
                },
                "errors": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "server.CityData": {
            "type": "object",
            "properties": {
                "city_id": {
                    "type": "integer"
                },
                "city_name": {
                    "type": "string"
                },
                "continent": {
                    "type": "string"
                },
                "country_3_code": {
                    "type": "string"
                }
            }
        },
        "server.TagsData": {
            "type": "object",
            "properties": {
                "air_quality": {
                    "type": "string"
                },
                "city_id": {
                    "type": "integer"
                },
                "city_size": {
                    "type": "string"
                },
                "cloud_coverage": {
                    "type": "string"
                },
                "daylight_hours": {
                    "type": "string"
                },
                "humidity": {
                    "type": "string"
                },
                "precipitation": {
                    "type": "string"
                },
                "temperature": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "description": "Authorization to access the API endpoints",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        },
        "BasicAuth": {
            "type": "basic"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.0.8",
	Host:             "city-tags-api.com",
	BasePath:         "/api/v0",
	Schemes:          []string{},
	Title:            "City Tags API",
	Description:      "This is an API that makes available different tags for worlwide cities",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
