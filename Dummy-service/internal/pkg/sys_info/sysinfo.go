package sys_info

import (
	"context"
	"github.com/unwisecode/over-the-horison-andress/platform/logging"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func GetRAMSample(ctx context.Context) (int64, int64) {
	var MemTotal, MemAvailable int64
	contents, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		logging.FromContext(ctx).Error(ctx, "error in GetRAMSample: "+err.Error())
		return 0, 0
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines[:len(lines)-1] {
		fields := strings.Fields(line)

		switch fields[0] {
		case "MemTotal:":
			MemTotal, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				// TODO: стандартизировать ошибки
				logging.FromContext(ctx).Warn(ctx, "error in GetRAMSample: "+err.Error())
			}
		case "MemAvailable:":
			MemAvailable, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				logging.FromContext(ctx).Warn(ctx, "error in GetRAMSample: "+err.Error())
			}
		}
	}
	return MemTotal, MemTotal - MemAvailable

}

func GetCPUSample(ctx context.Context) (map[string]uint64, map[string]uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	total := make(map[string]uint64)
	total_idle := make(map[string]uint64)
	if err != nil {
		logging.FromContext(ctx).Error(ctx, "error in GetCPUSample"+err.Error())
		return total, total_idle
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines[:len(lines)-1] {
		fields := strings.Fields(line)

		if strings.HasPrefix(fields[0], "cpu") {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					logging.FromContext(ctx).Warn(ctx, "error in GetCPUSample: "+err.Error())
				}
				total[fields[0]] += val // tally up all the numbers to get total ticks
				if i == 4 {             // idle is the 5th field in the cpu line
					total_idle[fields[0]] = val
				}
			}
		}
	}
	return total, total_idle
}

func CountCPUusage(ctx context.Context) map[string]float64 {
	total_0, total_idle_0 := GetCPUSample(ctx)
	time.Sleep(5 * time.Second)
	total_1, total_idle_1 := GetCPUSample(ctx)
	CPU_usage := make(map[string]float64)

	idleTicks := make(map[string]float64)
	totalTicks := make(map[string]float64)

	if len(total_0) == 0 || len(total_1) == 0 {
		logging.FromContext(ctx).Warn(ctx, "zero data from GetCPUSample")
		return CPU_usage
	}

	for key, _ := range total_0 {
		totalTicks[key] = float64(total_1[key] - total_0[key])
		idleTicks[key] = float64(total_idle_1[key] - total_idle_0[key])

		CPU_usage[key] = 100 * (totalTicks[key] - idleTicks[key]) / totalTicks[key]
	}

	return CPU_usage
}

func GetDISCSample(ctx context.Context, path string) (float64, float64) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		logging.FromContext(ctx).Error(ctx, "error in GetDISCSample"+err.Error())
		return 0, 0
	}
	disk_ALL := fs.Blocks * uint64(fs.Bsize)
	disk_FREE := fs.Bfree * uint64(fs.Bsize)
	disk_USED := disk_ALL - disk_FREE

	return float64(disk_ALL) / GB, float64(disk_USED) / GB
}
