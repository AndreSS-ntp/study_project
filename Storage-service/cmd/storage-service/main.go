package main

import (
	"context"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/app/storage-service"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/repository"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/service/system"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := alogger.NewLogger()
	baseCtx := alogger.WithLogger(context.Background(), logger)
	logger.Info(baseCtx, "Storage-service is running...")
	ctx := withGracefulShutdown(baseCtx)

	dbpool, err := pgxpool.New(ctx, config.DB_URL)
	if err != nil {
		logger.Error(ctx, "Unable to create connection pool: "+err.Error())
		return
	}
	defer dbpool.Close()

	serv := system.Service{}
	repo := repository.NewDataManager(dbpool)
	storageApp := storage_service.NewApp(&serv, repo)
	mux := http.NewServeMux()

	for pattern, command := range storageApp.Commands {
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
			return
		}
	}()

	<-ctx.Done()

	err_sd := server.Shutdown(context.Background())
	if err_sd != nil {
		logger.Error(ctx, "error shutting down server: "+err_sd.Error())
	}

	logger.Info(ctx, "Storage-service stopped.")
}

func withGracefulShutdown(baseCtx context.Context) context.Context {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exit
		alogger.FromContext(ctx).Warn(ctx, "Shutting down service...")
		cancel()
	}()
	return ctx
}
