package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	locator_service "github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/app/locator-service"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/config"
	http_client "github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/repository/http-client"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Locator-service is running...")

	mu := new(sync.RWMutex)

	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exit
		fmt.Println("Shutting down service...")
		cancel()
	}()
	var wg sync.WaitGroup

	logFile, err := os.OpenFile(config.PathLogs, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err_make := os.MkdirAll(config.PathLogs, os.ModePerm)
			if err_make != nil {
				err_make = fmt.Errorf("cant make dir with logs: %w", err_make)
				fmt.Println(err_make)
			}
			logFile, err = os.OpenFile(config.PathLogs, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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
				for id := 1; id < len(config.Adresses)+1; id++ {
					url := config.Adresses[int64(id)] + "/health"
					client := http_client.HttpClient{}
					service_health, j_err := client.GetSystem(url)

					if errors.Is(j_err, errors.New("not found")) {
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
					mu.Lock()
					_, err_w := logFile.Write(system_log)
					mu.Unlock()
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

	locatorApp := locator_service.NewApp(mu)
	mux := http.NewServeMux()

	for pattern, command := range locatorApp.Commands {
		mux.HandleFunc(pattern, command.Handler)
	}

	server := &http.Server{
		Addr:    config.IP_port,
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
