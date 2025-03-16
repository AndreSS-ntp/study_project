package main

import (
	"context"
	"fmt"
	locator_service "github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/app/locator-service"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/repository/file"
	http_client "github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/repository/http-client"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/service/former"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/service/logger"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

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

	repository := file.NewFileManager(config.PathLogs)
	form := former.NewService(repository)
	client := http_client.NewHttpClient(config.HTTPClientTimeout)
	logg := logger.NewLogger(config.Adresses, client, repository, form)
	locatorApp := locator_service.NewApp(form)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logg.Run(ctx)
	}()

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
