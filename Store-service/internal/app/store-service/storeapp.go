package store_service

import (
	"encoding/json"
	"fmt"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/domain"
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
		"/v1/users": Command{"POST - Создать нового пользователя, " +
			"GET - Получить коллекцию пользователей", s.User},
		"/v1/users/{id}": Command{"GET - Получить информацию о конкретном пользователе, " +
			"PUT/PATCH - Полное/Частичное обновление пользователя," +
			"DELETE - Удаление пользователя", s.UserManage},
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

// TODO: придумать более подходящий нейминг ручек store-service/переделать
func (a *App) User(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//
	case http.MethodPost:
		// create user
	}
}

func (a *App) UserManage(w http.ResponseWriter, r *http.Request) {}
