package main

import (
	"context"
	"fmt"
	store_service "github.com/unwisecode/over-the-horison-andress/Store-service/internal/app/store-service"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/service"
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

	serv := service.Service{}
	storeApp := store_service.NewApp(&serv)
	mux := http.NewServeMux()

	for pattern, command := range storeApp.Commands {
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

	fmt.Println("Dummy-service stopped.")
}
