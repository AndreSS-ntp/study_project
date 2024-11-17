package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}
var adresses = map[int64]string{
	1: "http://127.0.0.1:7001",
}

const ip_port string = "127.0.0.1:7000"

type System struct {
	num_CPU    int
	CPU_usage  map[string]float64
	RAM        int64
	RAM_used   int64
	DISC       float64
	DISC_used  float64
	GOMAXPROCS int
}

func serviceHEALTH(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		err = fmt.Errorf("Error while parsing ID from request: %w", err)
		fmt.Fprintln(w, err)
		return
	}

	url := adresses[id] + "/health"
	service_health := System{}

	j_err := getJson(url, &service_health)
	if j_err != nil {
		j_err = fmt.Errorf("Error while getting json with system info: %w", j_err)
		fmt.Fprintln(w, j_err)
		return
	}

	data, err := json.Marshal(service_health)
	if err != nil {
		err = fmt.Errorf("Error while marshling json: %w", err)
		fmt.Fprintln(w, err)
		return
	}
	
	fmt.Fprintf(w, string(data))
}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return fmt.Errorf("Error while getting URL from client: %w", err)
	}
	defer func() {
		def_err := r.Body.Close()
		if def_err != nil {
			def_err = fmt.Errorf("Error while closing response body: %w", def_err)
			fmt.Println(def_err)
		}
	}()
	return json.NewDecoder(r.Body).Decode(target)
}

func main() {
	http.HandleFunc("/serviceHEALTH/{id}", serviceHEALTH)

	err := http.ListenAndServe(ip_port, nil)
	if err != nil {
		err = fmt.Errorf("Cant ListenAndServe: %w", err)
		fmt.Println(err)
	}
}
