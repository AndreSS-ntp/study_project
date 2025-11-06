package main

import (
	"context"
	alogger "github.com/AndreSS-ntp/logger"
	locator_service "github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/app/locator-service"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/repository/file"
	http_client "github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/repository/http-client"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/service/former"
	logging "github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/service/logger"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	logger := alogger.NewLogger()
	baseCtx := alogger.WithLogger(context.Background(), logger)
	logger.Info(baseCtx, "Locator-service is running...")

	ctx, cancel := context.WithCancel(baseCtx)
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exit
		logger.Warn(ctx, "Shutting down service...")
		cancel()
	}()
	var wg sync.WaitGroup

	repository := file.NewFileManager(config.PathLogs)
	form := former.NewService(repository)
	client := http_client.NewHttpClient(config.HTTPClientTimeout)
	logg := logging.NewSysLogger(config.Adresses, client, repository, form)
	locatorApp := locator_service.NewApp(form)

	wg.Add(1)
	go func() {
		defer wg.Done()
		logg.Run(ctx)
	}()

	mux := http.NewServeMux()

	for pattern, command := range locatorApp.Commands {
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

	wg.Wait()
	logger.Info(ctx, "Locator-service stopped.")
}
