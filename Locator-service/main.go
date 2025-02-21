package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var myClient = &http.Client{Timeout: 10 * time.Second}
var adresses = map[int64]string{
	1: "http://172.17.0.1:7001",
	// 1: "http://127.0.0.1:7001", // для тестов без докера
}

const ip_port string = "0.0.0.0:7000"

var ErrNotFound = errors.New("not found")

type System struct {
	Num_CPU    int                `json:"num_cpu"`
	CPU_usage  map[string]float64 `json:"cpu_usage"`
	RAM        int64              `json:"ram"`
	RAM_used   int64              `json:"ram_used"`
	DISC       float64            `json:"disc"`
	DISC_used  float64            `json:"disc_used"`
	GOMAXPROCS int                `json:"gomaxprocs"`
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

	file, err_open := os.Open("../data/file-storage/system-data/logs")
	if err_open != nil {
		err_open = fmt.Errorf("500 - iternal server error: %w", err_open)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err_open.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}
	defer func() {
		err_close := file.Close()
		if err_close != nil {
			err_close = fmt.Errorf("cant close logFile: %w", err_close)
			fmt.Println(err_close)
		}
	}()

	in := bufio.NewReader(file)

	sb := strings.Builder{}
	for {
		line, err_read := in.ReadString('\n')
		if err_read != nil {
			if err_read == io.EOF {
				break
			} else {
				err_read = fmt.Errorf("500 - iternal server error: %w", err_read)
				w.WriteHeader(500)
				_, w_err := w.Write([]byte(err_read.Error()))
				if w_err != nil {
					w_err = fmt.Errorf("cant write a response: %w", w_err)
					fmt.Println(w_err)
				}
				return
			}
		}
		splited_line := strings.Split(line, " ")
		parsed_id, err_parse := strconv.ParseInt(splited_line[0], 10, 64)
		if err_parse != nil {
			err_parse = fmt.Errorf("500 - iternal server error: %w", err_parse)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err_parse.Error()))
			if w_err != nil {
				w_err = fmt.Errorf("cant write a response: %w", w_err)
				fmt.Println(w_err)
			}
			return
		}

		if parsed_id == id {
			sys_log := System{}
			err_unmarsh := json.Unmarshal([]byte(splited_line[3]), &sys_log)
			if err_unmarsh != nil {
				err_unmarsh = fmt.Errorf("500 - iternal server error: %w", err_unmarsh)
				w.WriteHeader(500)
				_, w_err := w.Write([]byte(err_unmarsh.Error()))
				if w_err != nil {
					w_err = fmt.Errorf("cant write a response: %w", w_err)
					fmt.Println(w_err)
				}
				return
			}
			i := -1
			var cpu_name string
			if sb.Len() == 0 {
				sb.WriteString("YYYY-MM-DD\thh-mm-ss\t")
				for range sys_log.CPU_usage {
					if i == -1 {
						cpu_name = "cpu"
						i++
					} else {
						cpu_name = "cpu" + strconv.Itoa(i)
						i++
					}
					sb.WriteString(cpu_name)
					sb.WriteString("\t")
				}
				sb.WriteString("ram\tram_used\tdisc\tdisc_used\tgomaxprocs\n")
			}

			bytes_to_grow := len(splited_line[1]) + len(splited_line[2]) + 2 +
				len(strconv.Itoa(sys_log.GOMAXPROCS)) + 1 +
				len(strconv.Itoa(sys_log.Num_CPU)) + 1 +
				len(strconv.FormatInt(sys_log.RAM, 10)) + 1 +
				len(strconv.FormatInt(sys_log.RAM_used, 10)) + 1 +
				len(strconv.FormatFloat(sys_log.DISC, 'f', -1, 64)) + 1 +
				len(strconv.FormatFloat(sys_log.DISC_used, 'f', -1, 64)) + 1
			for _, v := range sys_log.CPU_usage {
				bytes_to_grow += len(strconv.FormatFloat(v, 'f', -1, 64)) + 1
			}

			sb.Grow(bytes_to_grow)
			sb.WriteString(splited_line[1])
			sb.WriteString("\t")
			sb.WriteString(splited_line[2])
			sb.WriteString("\t")
			i = -1
			for range sys_log.CPU_usage {
				if i == -1 {
					cpu_name = "cpu"
					i++
				} else {
					cpu_name = "cpu" + strconv.Itoa(i)
					i++
				}
				sb.WriteString(strconv.FormatFloat(sys_log.CPU_usage[cpu_name], 'f', -1, 64))
				sb.WriteString("\t")
			}
			sb.WriteString(strconv.FormatInt(sys_log.RAM, 10))
			sb.WriteString("\t")
			sb.WriteString(strconv.FormatInt(sys_log.RAM_used, 10))
			sb.WriteString("\t")
			sb.WriteString(strconv.FormatFloat(sys_log.DISC, 'f', -1, 64))
			sb.WriteString("\t")
			sb.WriteString(strconv.FormatFloat(sys_log.DISC_used, 'f', -1, 64))
			sb.WriteString("\t")
			sb.WriteString(strconv.Itoa(sys_log.GOMAXPROCS))
			sb.WriteString("\n")

		}
	}

	w.WriteHeader(200)
	_, w_err := w.Write([]byte(sb.String()))
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
		if errors.Is(err, os.ErrNotExist) {
			err_make := os.MkdirAll("../data/file-storage/system-data", os.ModePerm)
			if err_make != nil {
				err_make = fmt.Errorf("cant make dir with logs: %w", err_make)
				fmt.Println(err_make)
			}
			logFile, err = os.OpenFile("../data/file-storage/system-data/logs", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				err = fmt.Errorf("cant open logFile: %w", err)
				fmt.Println(err)
			}
		} else {
			err = fmt.Errorf("cant open logFile: %w", err)
			fmt.Println(err)
		}
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

					sb := strings.Builder{}
					sb.Grow(len(data) + len(strconv.Itoa(id)) + 22) // 22 - кол-во байт рассчитанное на дату/время + " "x2 + "\n"
					sb.WriteString(strconv.Itoa(id))
					sb.WriteString(" ")
					sb.WriteString(time.Now().Format("2006-01-02 15:04:05"))
					sb.WriteString(" ")
					sb.WriteString(string(data))
					sb.WriteString("\n")

					system_log := []byte(sb.String())
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

	mux := http.NewServeMux()
	mux.HandleFunc("/service-health/{id}", serviceHealth)

	server := &http.Server{
		Addr:    ip_port,
		Handler: mux,
	}

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
