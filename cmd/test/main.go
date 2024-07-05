package main

import (
	"fmt"
	"net/http"
)

func main() {
	resp, _ := http.Get("http://localhost:8080/v0/cities/3838859")
	fmt.Print(resp)
}
