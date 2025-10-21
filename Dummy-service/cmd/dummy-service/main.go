package main

import (
	"context"
	alogger "github.com/AndreSS-ntp/logger"
	dummy_service "github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/app/dummy-service"
	"github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/service"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := alogger.NewLogger()
	baseCtx := alogger.WithLogger(context.Background(), logger)
	logger.Info(baseCtx, "Dummy-service is running...")
	ctx, cancel := context.WithCancel(baseCtx)
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exit
		logger.Warn(ctx, "Shutting down service...")
		cancel()
	}()

	serv := service.Service{}
	dummyApp := dummy_service.NewApp(&serv)
	mux := http.NewServeMux()

	for pattern, command := range dummyApp.Commands {
		mux.HandleFunc(pattern, withLogger(logger, command.Handler))
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

	logger.Info(ctx, "Dummy-service stopped.")
}

func withLogger(logger alogger.Logger, handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := alogger.WithLogger(r.Context(), logger)
		handler(w, r.WithContext(ctx))
	}
}
