package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}
var adresses = map[int64]string{
	1: "http://172.17.0.1:7001",
}

const ip_port string = "0.0.0.0:7000"

var ErrNotFound = errors.New("not found")

type System struct {
	num_CPU    int
	CPU_usage  map[string]float64
	RAM        int64
	RAM_used   int64
	DISC       float64
	DISC_used  float64
	GOMAXPROCS int
}

func serviceHealth(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		err = fmt.Errorf("500 - iternal server error: %w", err)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}

	url := adresses[id] + "/health"
	service_health := System{}

	j_err := getJson(url, &service_health)
	if errors.Is(j_err, ErrNotFound) {
		j_err = fmt.Errorf("404 - not found")
		w.WriteHeader(404)
		_, w_err := w.Write([]byte(j_err.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}

	data, err := json.Marshal(service_health)
	if err != nil {
		err = fmt.Errorf("500 - iternal server error: %w", err)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}

	w.WriteHeader(200)
	_, w_err := w.Write(data)
	if w_err != nil {
		w_err = fmt.Errorf("cant write a response: %w", w_err)
		fmt.Println(w_err)
	}
}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return fmt.Errorf("error while getting URL from client: %w", ErrNotFound)
	}
	defer func() {
		def_err := r.Body.Close()
		if def_err != nil {
			def_err = fmt.Errorf("error while closing response body: %w", def_err)
			fmt.Println(def_err)
		}
	}()
	return json.NewDecoder(r.Body).Decode(target)
}

func main() {
	fmt.Println("Locator-service is now running...")
	http.HandleFunc("/serviceHealth/{id}", serviceHealth)

	err := http.ListenAndServe(ip_port, nil)
	if err != nil {
		err = fmt.Errorf("cant ListenAndServe: %w", err)
		fmt.Println(err)
	}
}
