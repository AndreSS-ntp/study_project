package logger

import (
	"context"
	"errors"
	"fmt"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/domain"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/service/former"
	"time"
)

type Logger struct {
	Adresses    map[int]string
	GetSystemer GetSystemer
	Repository  former.Repository
	Formater    Formater
}

type GetSystemer interface {
	GetSystem(ctx context.Context, url string) (*domain.System, error)
}

type Formater interface {
	GetLogFormat(ctx context.Context, sys_logs map[int]*domain.System) []string
}

func NewLogger(adresses map[int]string, getsystemer GetSystemer, repository former.Repository, formater Formater) *Logger {
	return &Logger{adresses, getsystemer, repository, formater}
}

func (l *Logger) getLogs(ctx context.Context) map[int]*domain.System {
	sys_logs := make(map[int]*domain.System, len(config.Adresses))
	for id := 1; id < len(config.Adresses)+1; id++ {
		url := config.Adresses[id] + "/health"

		service_health, j_err := l.GetSystemer.GetSystem(ctx, url)

		if errors.Is(j_err, errors.New("not found")) {
			alogger.FromContext(ctx).Error(ctx, fmt.Sprintf("404 - not found (service_id: %d)", id))
			continue
		}
		sys_logs[id] = service_health
	}
	return sys_logs
}

func (l *Logger) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			alogger.FromContext(ctx).Warn(ctx, "Logging stopped: "+ctx.Err().Error())
			return
		default:
			sys_logs := l.getLogs(ctx)
			str_logs := l.Formater.GetLogFormat(ctx, sys_logs)
			for _, log := range str_logs {
				err := l.Repository.WriteLog(ctx, log)
				if err != nil {
					alogger.FromContext(ctx).Error(ctx, "error writing log: "+err.Error())
					return
				}
			}
			alogger.FromContext(ctx).Info(ctx, "Log done.")

			time.Sleep(5 * time.Second)
		}
	}
}
