package storage_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/govalues/money"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/domain"
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
	CreateItem(ctx context.Context, item *domain.Item) (domain.Item, error)
	UpdateProduct(ctx context.Context, item *domain.Item) (domain.Item, error)
	GetItemBySKU(ctx context.Context, sku uint64) (*domain.Item, error)
	DeleteItem(ctx context.Context, sku uint64) error
	ListItems(ctx context.Context, limit, offset int) ([]domain.Item, error)
}

// для десериализации
type ItemParse struct {
	SKU      uint64 `json:"sku"`
	Name     string `json:"name"`
	Price    string `json:"price"`
	Quantity int    `json:"quantity"`
}

func NewApp(h GetSystemer, r Repository) *App {
	s := App{}
	var commands = map[string]Command{
		"GET /help":             Command{"Список команд.", s.Help},
		"GET /health":           Command{"Вернуть состояние сервиса и данные о системе сервера.", s.Health},
		"GET /v1/item/{sku}":    Command{"Получить предмет по SKU", s.GetItem},
		"DELETE /v1/item/{sku}": Command{"Удалить предмет по SKU", s.DeleteItem},
		"GET /v1/items":         Command{"Получить список всех товаров (параметры для пагинации: limit, offset)", s.GetItems},
		"POST /v1/item":         Command{"Создать новый товар", s.CreateItem},
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

func (a *App) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sku, err := strconv.ParseUint(r.PathValue("sku"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid sku"}`, 400)
		/*err = fmt.Errorf("invalid item sku: %w", err)
		w.WriteHeader(400)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}*/
		return
	}
	item, err := a.Repository.GetItemBySKU(ctx, sku)
	if err != nil {
		http.Error(w, `{"error":"item not found"}`, 404)
		/*err = fmt.Errorf("item not found: %w", err)
		w.WriteHeader(404)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}*/
		return
	}
	data, err := json.Marshal(item)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, 404)
		/*err = fmt.Errorf("internal server error: %w", err)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}*/
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, w_err := w.Write(data)
	if w_err != nil {
		alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
	}
}

func (a *App) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sku, err := strconv.ParseUint(r.PathValue("sku"), 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid sku"}`, 400)
		/*err = fmt.Errorf("invalid user sku: %w", err)
		w.WriteHeader(400)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}*/
		return
	}

	err = a.Repository.DeleteItem(ctx, sku)
	if err != nil {
		if errors.Is(err, fmt.Errorf("item not found")) {
			http.Error(w, `{"error":"item not found"}`, 404)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, 500)
		/*err = fmt.Errorf("internal server error: %w", err)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}*/
		return
	}
	w.WriteHeader(204)
}

func (a *App) CreateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var item ItemParse
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		err = fmt.Errorf("internal server error: %w", err)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		return
	}

	amount, err := money.ParseAmount("RUB", item.Price)
	if err != nil {
		err = fmt.Errorf("invalid price format: %w", err)
		w.WriteHeader(400)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		return
	}

	newItem := domain.Item{
		SKU:      item.SKU,
		Name:     item.Name,
		Price:    amount,
		Quantity: item.Quantity,
	}

	createdItem, err := a.Repository.CreateItem(ctx, &newItem)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, `{"error":"sku already exists"}`, http.StatusConflict)
			return
		}

		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(createdItem)
	if err != nil {
		err = fmt.Errorf("internal server error: %w", err)
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
}

func (a *App) GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	items, err := a.Repository.ListItems(ctx, limit, offset)
	if err != nil {
		err = fmt.Errorf("internal server error: %w", err)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		return
	}

	data, err := json.Marshal(items)
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
}

func (a *App) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sku, err := strconv.ParseUint(r.PathValue("sku"), 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid user sku: %w", err)
		w.WriteHeader(400)
		_, w_err := w.Write([]byte(err.Error()))
		if w_err != nil {
			alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
		}
		return
	}

	var itemToUpdate ItemParse
	err = json.NewDecoder(r.Body).Decode(&itemToUpdate)
	if err != nil {
		http.Error(w, `{"error":"invalid json"}`, 400)
		return
	}

	amount, err := money.ParseAmount("RUB", itemToUpdate.Price)
	if err != nil {
		http.Error(w, `{"error":"invalid price format"}`, 400)
		return
	}

	updatedItem := domain.Item{
		SKU:      sku,
		Name:     itemToUpdate.Name,
		Price:    amount,
		Quantity: itemToUpdate.Quantity,
	}

	updated, err := a.Repository.UpdateProduct(ctx, &updatedItem)
	if err != nil {
		if errors.Is(err, fmt.Errorf("item not found")) {
			http.Error(w, `{"error":"item not found"}`, http.StatusNotFound)
			return
		}

		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(updated)
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
}
