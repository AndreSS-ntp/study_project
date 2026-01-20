package system

import (
	"context"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/domain"
	"github.com/unwisecode/over-the-horison-andress/Storage-service/internal/pkg/sys_info"
	"runtime"
)

type Service struct {
	Repository Repository
}

type Repository interface {
	CreateItem(ctx context.Context, item *domain.Item) (*domain.ItemToSend, error)
	UpdateProduct(ctx context.Context, item *domain.Item) (*domain.ItemToSend, error)
	GetItemBySKU(ctx context.Context, sku uint64) (*domain.ItemToSend, error)
	DeleteItem(ctx context.Context, sku uint64) error
	ListItems(ctx context.Context, limit, offset int) ([]*domain.ItemToSend, error)
}

func NewService(repository Repository) *Service {
	return &Service{repository}
}

func (s *Service) CreateItem(ctx context.Context, item *domain.Item) (*domain.ItemToSend, error) {
	return s.Repository.CreateItem(ctx, item)
}

func (s *Service) UpdateProduct(ctx context.Context, item *domain.Item) (*domain.ItemToSend, error) {
	return s.Repository.UpdateProduct(ctx, item)
}

func (s *Service) GetItemBySKU(ctx context.Context, sku uint64) (*domain.ItemToSend, error) {
	return s.Repository.GetItemBySKU(ctx, sku)
}

func (s *Service) DeleteItem(ctx context.Context, sku uint64) error {
	return s.Repository.DeleteItem(ctx, sku)
}

func (s *Service) ListItems(ctx context.Context, limit, offset int) ([]*domain.ItemToSend, error) {
	return s.Repository.ListItems(ctx, limit, offset)
}

func (*Service) GetSystem(ctx context.Context) *domain.System {
	s := domain.System{}
	s.Num_CPU = runtime.NumCPU()
	s.CPU_usage = sys_info.CountCPUusage(ctx)
	s.RAM, s.RAM_used = sys_info.GetRAMSample(ctx)
	s.DISC, s.DISC_used = sys_info.GetDISCSample(ctx, "/")
	s.GOMAXPROCS = runtime.GOMAXPROCS(0)
	return &s
}
