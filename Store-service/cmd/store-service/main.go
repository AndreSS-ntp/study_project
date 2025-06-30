package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	store_service "github.com/unwisecode/over-the-horison-andress/Store-service/internal/app/store-service"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/repository"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/service/system"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Store-service is running...")
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	dbpool, err := pgxpool.New(ctx, config.DB_URL)
	if err != nil {
		err = fmt.Errorf("Unable to create connection pool: %v\n", err)
		fmt.Println(err)
		cancel()
	}
	defer dbpool.Close()

	go func() {
		<-exit
		fmt.Println("Shutting down service...")
		cancel()
	}()

	serv := system.Service{}
	repo := repository.NewDataManager(dbpool)
	storeApp := store_service.NewApp(&serv, repo)
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
