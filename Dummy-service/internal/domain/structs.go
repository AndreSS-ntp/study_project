package domain

import (
	"github.com/unwisecode/over-the-horison-andress/Dummy-service/internal/pkg/sys_info"
	"runtime"
)

type System struct {
	Num_CPU    int                `json:"num_cpu"`
	CPU_usage  map[string]float64 `json:"cpu_usage"`
	RAM        int64              `json:"ram"`
	RAM_used   int64              `json:"ram_used"`
	DISC       float64            `json:"disc"`
	DISC_used  float64            `json:"disc_used"`
	GOMAXPROCS int                `json:"gomaxprocs"`
}

func NewSystem() System {
	s := System{}
	s.Num_CPU = runtime.NumCPU()
	s.CPU_usage = sys_info.CountCPUusage()
	s.RAM, s.RAM_used = sys_info.GetRAMSample()
	s.DISC, s.DISC_used = sys_info.GetDISCSample("/")
	s.GOMAXPROCS = runtime.GOMAXPROCS(0)
	return s
}
