package main

import (
	"city-tags-api-iac/internal/aws_lib"
	"encoding/json"
	"fmt"
)

func main() {
	ssm := aws_lib.NewSSM()
	param := ssm.GetParam("/city-tags-api/dev/db", true)
	dbParam := map[string]string{}
	json.Unmarshal([]byte(param), &dbParam)
	fmt.Print(dbParam)
}
