package service

import (
	"encoding/json"
	"fmt"
	"github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/domain"
	"net/http"
)

var commands_info = map[string]string{
	"/help":   "Список команд.",
	"/health": "Вернуть состояние сервиса и данные о системе сервера.",
}

type Service struct {
	Commands map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewService() *Service {
	s := Service{}
	var commands = map[string]func(w http.ResponseWriter, r *http.Request){
		"/health": s.Health,
		"/help":   s.Help,
	}
	s.Commands = commands
	return &s
}

func (*Service) Health(w http.ResponseWriter, r *http.Request) {
	system := domain.NewSystem()
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

func (*Service) Help(w http.ResponseWriter, r *http.Request) {
	message := ""
	for key, value := range commands_info {
		message += key + " - " + value + "\n"
	}
	w.WriteHeader(http.StatusOK)
	_, w_err := w.Write([]byte(message))
	if w_err != nil {
		w_err = fmt.Errorf("cant write a response: %w", w_err)
		fmt.Println(w_err)
	}
}
