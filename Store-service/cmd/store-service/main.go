package main

import (
	"context"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	store_service "github.com/unwisecode/over-the-horison-andress/Store-service/internal/app/store-service"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/repository"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/service/system"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := alogger.NewLogger()
	baseCtx := alogger.WithLogger(context.Background(), logger)
	logger.Info(baseCtx, "Store-service is running...")
	ctx, cancel := context.WithCancel(baseCtx)
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	dbpool, err := pgxpool.New(ctx, config.DB_URL)
	if err != nil {
		logger.Error(ctx, "Unable to create connection pool: "+err.Error())
		cancel()
	}
	defer dbpool.Close()

	go func() {
		<-exit
		logger.Warn(ctx, "Shutting down service...")
		cancel()
	}()

	serv := system.Service{}
	repo := repository.NewDataManager(dbpool)
	storeApp := store_service.NewApp(&serv, repo)
	mux := http.NewServeMux()

	for pattern, command := range storeApp.Commands {
		mux.HandleFunc(pattern, alogger.HandlerWithLogger(logger, command.Handler))
	}

	server := &http.Server{
		Addr:    config.IP_port,
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		err_las := server.ListenAndServe()
		if err_las != nil {
			logger.Error(ctx, "HTTP server error: "+err_las.Error())
			cancel()
		}
	}()

	<-ctx.Done()

	err_sd := server.Shutdown(context.Background())
	if err_sd != nil {
		logger.Error(ctx, "error shutting down server: "+err_sd.Error())
	}

	logger.Info(ctx, "Store-service stopped.")
}
