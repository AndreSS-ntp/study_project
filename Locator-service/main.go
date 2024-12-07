package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
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
	fmt.Println("Locator-service is running...")

	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exit
		fmt.Println("Shutting down service...")
		cancel()
	}()

	var wg sync.WaitGroup

	logFile, err := os.OpenFile("../data/file-storage/system-data/logs", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = fmt.Errorf("cant open logFile: %w", err)
		fmt.Println(err)
	}
	defer func() {
		err_close := logFile.Close()
		if err_close != nil {
			err_close = fmt.Errorf("cant close logFile: %w", err_close)
			fmt.Println(err_close)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err != nil {
			return // Если LogFile не открылся
		}
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Logging stopped:", ctx.Err().Error())
				return
			default:
				for id := 1; id < len(adresses)+1; id++ {
					url := adresses[int64(id)] + "/health"
					service_health := System{}

					j_err := getJson(url, &service_health)
					if errors.Is(j_err, ErrNotFound) {
						j_err = fmt.Errorf("404 - not found")
						fmt.Println(j_err)
						continue
					}

					data, err := json.Marshal(service_health)
					if err != nil {
						err = fmt.Errorf("error occured: %w", err)
						fmt.Println(err)
					}
					system_log := []byte(fmt.Sprintf("%d %d.%d.%d %d:%d:%d %s\n", id, time.Now().Year(), time.Now().Month(), time.Now().Day(),
						time.Now().Hour(), time.Now().Minute(), time.Now().Second(), data))
					_, err_w := logFile.Write(system_log)
					if err_w != nil {
						err = fmt.Errorf("write error occured: %w", err)
						fmt.Println(err)
					}
					fmt.Println("Log")
				}
				time.Sleep(5 * time.Second)
			}
		}
	}()

	server := &http.Server{Addr: ip_port, Handler: nil}

	http.HandleFunc("/serviceHealth/{id}", serviceHealth)

	go func() {
		err_las := server.ListenAndServe()
		if err_las != nil {
			err_las = fmt.Errorf("HTTP server error: %w", err_las)
			fmt.Println(err_las)
			cancel()
		}
	}()

	<-ctx.Done()

	err_sd := server.Shutdown(context.Background())
	if err_sd != nil {
		err_sd = fmt.Errorf("error shutting down server: %v", err_sd)
		fmt.Println(err_sd)
	}

	wg.Wait()
	fmt.Println("Locator-service stopped.")

}
