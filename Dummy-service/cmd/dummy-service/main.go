package main

import (
	"context"
	"fmt"
	dummy_service "github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/app/dummy-service"
	"github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Dummy-service is running...")
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exit
		fmt.Println("Shutting down service...")
		cancel()
	}()

	dummyApp := dummy_service.NewApp()
	mux := http.NewServeMux()

	for pattern, handler := range dummyApp.Commands {
		mux.HandleFunc(pattern, handler)
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

	fmt.Println("Dummy-service stopped.")
}
