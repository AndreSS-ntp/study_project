package storage_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/govalues/money"
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
	CreateItem(ctx context.Context, item *domain.Item) (*domain.ItemToSend, error)
	UpdateProduct(ctx context.Context, item *domain.Item) (*domain.ItemToSend, error)
	GetItemBySKU(ctx context.Context, sku uint64) (*domain.ItemToSend, error)
	DeleteItem(ctx context.Context, sku uint64) error
	ListItems(ctx context.Context, limit, offset int) ([]*domain.ItemToSend, error)
}

// Объект item для десериализации из запроса
type itemParse struct {
	SKU      uint64 `json:"sku"`
	Name     string `json:"name"`
	Price    string `json:"price"`
	Quantity int    `json:"quantity"`
}

type ErrorResponse struct {
	ErrMsg string `json:"error"`
}

func NewApp(h GetSystemer, r Repository) *App {
	s := App{}
	var commands = map[string]Command{
		"GET /help":             Command{"Список команд.", s.Help},
		"GET /health":           Command{"Вернуть состояние сервиса и данные о системе сервера.", s.Health},
		"GET /v1/item/{sku}":    Command{"Получить предмет по SKU", s.GetItem},
		"PUT /v1/item/{sku}":    Command{"Обновить товар по SKU", s.UpdateItem},
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
		sendError(ctx, w, err.Error(), 500)
		return
	}
	sendOk(ctx, w, data, 200)
}

func (a *App) Help(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	message := ""
	for pattern, command := range a.Commands {
		message += pattern + " - " + command.Description + "\n"
	}
	sendOk(ctx, w, []byte(message), 200)
}

func (a *App) GetItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sku, err := strconv.ParseUint(r.PathValue("sku"), 10, 64)
	if err != nil {
		sendError(ctx, w, "invalid sku", 400)
		return
	}
	item, err := a.Repository.GetItemBySKU(ctx, sku)
	if err != nil {
		sendError(ctx, w, "item not found", 404)
		return
	}

	data, err := json.Marshal(item)

	if err != nil {
		sendError(ctx, w, "internal server error", 500)
		return
	}
	sendOk(ctx, w, data, 200)
}

func (a *App) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sku, err := strconv.ParseUint(r.PathValue("sku"), 10, 64)
	if err != nil {
		sendError(ctx, w, "invalid sku", 400)
		return
	}

	err = a.Repository.DeleteItem(ctx, sku)
	if err != nil {
		if errors.Is(err, fmt.Errorf("item not found")) {
			sendError(ctx, w, "item not found", 404)
			return
		}
		sendError(ctx, w, "internal server error", 500)
		return
	}
	w.WriteHeader(204)
}

func (a *App) CreateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var item itemParse
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		sendError(ctx, w, "invalid json", 400)
		return
	}

	amount, err := money.ParseAmount("RUB", item.Price)
	if err != nil {
		sendError(ctx, w, "internal price format", 400)
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
		if errors.Is(err, domain.ErrAlreadyExists) {
			sendError(ctx, w, domain.ErrAlreadyExists.Error(), 409)
			return
		}
		sendError(ctx, w, "internal server error", 500)
		return
	}

	data, err := json.Marshal(createdItem)
	if err != nil {
		sendError(ctx, w, "internal server error", 500)
		return
	}

	sendOk(ctx, w, data, 201)
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
		sendError(ctx, w, "internal server error", 500)
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		sendError(ctx, w, "internal server error", 500)
		return
	}

	sendOk(ctx, w, data, 200)
}

func (a *App) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sku, err := strconv.ParseUint(r.PathValue("sku"), 10, 64)
	if err != nil {
		sendError(ctx, w, "invalid sku", 400)
		return
	}

	var itemToUpdate itemParse
	err = json.NewDecoder(r.Body).Decode(&itemToUpdate)
	if err != nil {
		sendError(ctx, w, "invalid json", 400)
		return
	}

	amount, err := money.ParseAmount("RUB", itemToUpdate.Price)
	if err != nil {
		sendError(ctx, w, "invalid price format", 400)
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
			sendError(ctx, w, "item not found", 404)
			return
		}
		sendError(ctx, w, "internal server error", 500)
		return
	}

	data, err := json.Marshal(updated)
	if err != nil {
		sendError(ctx, w, "internal server error", 500)
		return
	}

	sendOk(ctx, w, data, 200)
}

func sendError(ctx context.Context, w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	w_err := json.NewEncoder(w).Encode(ErrorResponse{msg})
	if w_err != nil {
		alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
	}
}

func sendOk(ctx context.Context, w http.ResponseWriter, data []byte, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, w_err := w.Write(data)
	if w_err != nil {
		alogger.FromContext(ctx).Error(ctx, "cant write a response: "+w_err.Error())
	}
}
