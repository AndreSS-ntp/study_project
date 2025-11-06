package system

import (
	"context"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/domain"
	"github.com/unwisecode/over-the-horison-andress/Store-service/internal/pkg/sys_info"
	"runtime"
)

type Service struct{}

func (*Service) GetSystem(ctx context.Context) *domain.System {
	s := domain.System{}
	s.Num_CPU = runtime.NumCPU()
	s.CPU_usage = sys_info.CountCPUusage(ctx)
	s.RAM, s.RAM_used = sys_info.GetRAMSample(ctx)
	s.DISC, s.DISC_used = sys_info.GetDISCSample(ctx, "/")
	s.GOMAXPROCS = runtime.GOMAXPROCS(0)
	return &s
}
