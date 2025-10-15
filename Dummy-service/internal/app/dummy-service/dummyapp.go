package dummy_service

import (
	"context"
	"encoding/json"
	"github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/domain"
	"github.com/unwisecode/over-the-horison-andress/platform/logging"
	"net/http"
)

type Command struct {
	Description string
	Handler     func(http.ResponseWriter, *http.Request)
}

type App struct {
	Commands map[string]Command
	Service  GetSystemer
}

type GetSystemer interface {
	GetSystem(ctx context.Context) *domain.System
}

func NewApp(h GetSystemer) *App {
	s := App{}
	var commands = map[string]Command{
		"/help":   Command{"Список команд.", s.Help},
		"/health": Command{"Вернуть состояние сервиса и данные о системе сервера.", s.Health},
	}
	s.Commands = commands
	s.Service = h
	return &s
}

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	system := a.Service.GetSystem(ctx)
	data, err := json.Marshal(system)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			logging.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, w_err := w.Write(data)
	if w_err != nil {
		logging.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
	}
}

func (a *App) Help(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	message := ""
	for pattern, command := range a.Commands {
		message += pattern + " - " + command.Description + "\n"
	}
	w.WriteHeader(http.StatusOK)
	_, w_err := w.Write([]byte(message))
	if w_err != nil {
		logging.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
	}
}
