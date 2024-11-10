package main

import (
	"encoding/json"
	"fmt"
	"log"
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
		fmt.Fprintf(w, "wrong service id")
		return
	}

	url := adresses[id] + "/health"
	service_health := System{}
	getJson(url, &service_health)
	data, err := json.Marshal(service_health)
	if err != nil {
		fmt.Fprintf(w, "ERROR: marshling ", err)
	}
	fmt.Fprintf(w, string(data))
}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return nil
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func main() {
	http.HandleFunc("/serviceHEALTH/{id}", serviceHEALTH)

	err := http.ListenAndServe(ip_port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
