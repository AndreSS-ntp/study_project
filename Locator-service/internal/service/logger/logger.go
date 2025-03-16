package logger

import (
	"context"
	"errors"
	"fmt"
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
	GetSystem(url string) (*domain.System, error)
}

type Formater interface {
	GetLogFormat(sys_logs map[int]*domain.System) []string
}

func NewLogger(adresses map[int]string, getsystemer GetSystemer, repository former.Repository, formater Formater) *Logger {
	return &Logger{adresses, getsystemer, repository, formater}
}

func (l *Logger) getLogs() map[int]*domain.System {
	sys_logs := make(map[int]*domain.System, len(config.Adresses))
	for id := 1; id < len(config.Adresses)+1; id++ {
		url := config.Adresses[id] + "/health"

		service_health, j_err := l.GetSystemer.GetSystem(url)

		if errors.Is(j_err, errors.New("not found")) {
			j_err = fmt.Errorf("404 - not found (service_id: %d)", id)
			fmt.Println(j_err)
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
			fmt.Println("Logging stopped:", ctx.Err().Error())
			return
		default:
			sys_logs := l.getLogs()
			str_logs := l.Formater.GetLogFormat(sys_logs)
			for _, log := range str_logs {
				err := l.Repository.WriteLog(log)
				if err != nil {
					err = fmt.Errorf("error writing log: %w", err)
					fmt.Println(err)
					return
				}
			}
			fmt.Println("Log")

			time.Sleep(5 * time.Second)
		}
	}
}
