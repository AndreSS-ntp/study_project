package locator_service

import (
	"fmt"
	"net/http"
	"strconv"
)

type Command struct {
	Description string
	Handler     func(http.ResponseWriter, *http.Request)
}

type App struct {
	Commands map[string]Command
	GetCSVer GetCSVer
}

type GetCSVer interface {
	GetCSV(id int64) ([]byte, error)
}

func NewApp(getcsver GetCSVer) *App {
	a := App{}
	var commands = map[string]Command{
		"/service-health/{id}": Command{"Состояние сервиса под ID.", a.ServiceHealth},
	}
	a.Commands = commands
	a.GetCSVer = getcsver
	return &a
}

func (a *App) ServiceHealth(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		err = fmt.Errorf("500 - internal server error: %w", err)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}
	logs_csv, err_getcsv := a.GetCSVer.GetCSV(id)
	if err_getcsv != nil {
		err = fmt.Errorf("500 - internal server error: %w", err_getcsv)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err_getcsv.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/csv")
	_, w_err := w.Write(logs_csv)
	if w_err != nil {
		w_err = fmt.Errorf("cant write a response: %w", w_err)
		fmt.Println(w_err)
	}
}
