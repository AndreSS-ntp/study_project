package dummy_service

import (
	"encoding/json"
	"fmt"
	"github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/domain"
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
	GetSystem() *domain.System
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
	system := a.Service.GetSystem()
	data, err := json.Marshal(system)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, w_err := w.Write(data)
	if w_err != nil {
		w_err = fmt.Errorf("cant write a response: %w", w_err)
		fmt.Println(w_err)
	}
}

func (a *App) Help(w http.ResponseWriter, r *http.Request) {
	message := ""
	for pattern, command := range a.Commands {
		message += pattern + " - " + command.Description + "\n"
	}
	w.WriteHeader(http.StatusOK)
	_, w_err := w.Write([]byte(message))
	if w_err != nil {
		w_err = fmt.Errorf("cant write a response: %w", w_err)
		fmt.Println(w_err)
	}
}
