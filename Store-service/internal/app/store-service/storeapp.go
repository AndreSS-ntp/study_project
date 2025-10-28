package store_service

import (
	"context"
	"encoding/json"
	"fmt"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/domain"
	"net/http"
	"strconv"
)

type Command struct {
	Description string
	Handler     func(http.ResponseWriter, *http.Request)
}

type App struct {
	Commands   map[string]Command
	Service    GetSystemer
	Repository Repository
}

type GetSystemer interface {
	GetSystem(ctx context.Context) *domain.System
}

type Repository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
	GetUserById(ctx context.Context, id int) (*domain.User, error)
	DeleteUser(ctx context.Context, id int) error
	ListUsers(ctx context.Context) ([]domain.User, error)
}

func NewApp(h GetSystemer, r Repository) *App {
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
	s.Repository = r
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
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, w_err := w.Write(data)
	if w_err != nil {
		alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
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
		alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
	}
}

// TODO: придумать более подходящий нейминг хендлеров store-service/переделать
func (a *App) User(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet: // list
		users, err := a.Repository.ListUsers(r.Context())
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}

		data, err := json.Marshal(users)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		_, w_err := w.Write(data)
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
	case http.MethodPost: // create user
		var user domain.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}

		err = a.Repository.CreateUser(r.Context(), &user)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}

		data, err := json.Marshal(user)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, w_err := w.Write(data)
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
	default:
		err := fmt.Errorf("405 - method not allowed")
		w.WriteHeader(405)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
	}
}

func (a *App) UserManage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		err = fmt.Errorf("400 - bad request, invalid user id: %w", err)
		w.WriteHeader(400)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		return
	}

	switch r.Method {
	case http.MethodGet: // выдать пользователя под индексом id
		user, err := a.Repository.GetUserById(ctx, id)
		if err != nil {
			err = fmt.Errorf("404 - user not found: %w", err)
			w.WriteHeader(404)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}

		data, err := json.Marshal(user)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, w_err := w.Write(data)
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		// TODO: store-service реализовать метод patch
	case http.MethodPut: // обновить пользователя (целиком)
		var user domain.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}
		user.ID = id

		err = a.Repository.UpdateUser(ctx, &user)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}
		w.WriteHeader(204)
	case http.MethodDelete: // удалить пользователя
		err := a.Repository.DeleteUser(ctx, id)
		if err != nil {
			err = fmt.Errorf("500 - internal server error: %w", err)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err.Error()))
			if w_err != nil {
				alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
			}
			return
		}
		w.WriteHeader(204)
	default:
		err := fmt.Errorf("405 - method not allowed")
		w.WriteHeader(405)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
	}
}
