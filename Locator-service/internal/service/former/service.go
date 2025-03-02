package former

import (
	"encoding/json"
	"fmt"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/domain"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	Repository Repository
}

type Repository interface {
	GetLogsById(id int64) ([]domain.System, [][]string, error)
	WriteLog(log string) error
}

func NewService(repository Repository) *Service {
	return &Service{repository}
}

func (s *Service) GetLogFormat(sys_logs map[int]*domain.System) []string {
	str_logs := make([]string, 0, len(sys_logs))
	for id, log := range sys_logs {
		data, err := json.Marshal(log)
		if err != nil {
			err = fmt.Errorf("error occured: %w", err)
			fmt.Println(err)
		}

		sb := strings.Builder{}
		sb.Grow(len(data) + len(strconv.Itoa(id)) + 22) // 22 - кол-во байт рассчитанное на дату/время + " "x2 + "\n"
		sb.WriteString(strconv.Itoa(id))
		sb.WriteString(" ")
		sb.WriteString(time.Now().Format("2006-01-02 15:04:05"))
		sb.WriteString(" ")
		sb.WriteString(string(data))
		sb.WriteString("\n")

		str_logs = append(str_logs, sb.String())
	}
	return str_logs
}

func (s *Service) GetCSV(id int64) ([]byte, error) {
	logs, timestamps, err := s.Repository.GetLogsById(id)
	if err != nil {
		return nil, err
	}
	sb := strings.Builder{}
	var i = -1
	var cpu_name string
	for j, sys_log := range logs {
		if sb.Len() == 0 {
			sb.WriteString("YYYY-MM-DD\thh-mm-ss\tnum_cpu\t")
			for range sys_log.CPU_usage {
				if i == -1 {
					cpu_name = "cpu"
					i++
				} else {
					cpu_name = "cpu" + strconv.Itoa(i)
					i++
				}
				sb.WriteString(cpu_name)
				sb.WriteString("\t")
			}
			sb.WriteString("ram\tram_used\tdisc\tdisc_used\tgomaxprocs\n")
		}
		bytes_to_grow := len(timestamps[j][0]) + len(timestamps[j][1]) + 2 +
			len(strconv.Itoa(sys_log.GOMAXPROCS)) + 1 +
			len(strconv.Itoa(sys_log.Num_CPU)) + 1 +
			len(strconv.FormatInt(sys_log.RAM, 10)) + 1 +
			len(strconv.FormatInt(sys_log.RAM_used, 10)) + 1 +
			len(strconv.FormatFloat(sys_log.DISC, 'f', -1, 64)) + 1 +
			len(strconv.FormatFloat(sys_log.DISC_used, 'f', -1, 64)) + 1
		for _, v := range sys_log.CPU_usage {
			bytes_to_grow += len(strconv.FormatFloat(v, 'f', -1, 64)) + 1
		}

		sb.Grow(bytes_to_grow)
		sb.WriteString(timestamps[j][0])
		sb.WriteString("\t")
		sb.WriteString(timestamps[j][1])
		sb.WriteString("\t")
		sb.WriteString(strconv.Itoa(sys_log.Num_CPU))
		sb.WriteString("\t")
		i = -1
		for range sys_log.CPU_usage {
			if i == -1 {
				cpu_name = "cpu"
				i++
			} else {
				cpu_name = "cpu" + strconv.Itoa(i)
				i++
			}
			sb.WriteString(strconv.FormatFloat(sys_log.CPU_usage[cpu_name], 'f', -1, 64))
			sb.WriteString("\t")
		}
		sb.WriteString(strconv.FormatInt(sys_log.RAM, 10))
		sb.WriteString("\t")
		sb.WriteString(strconv.FormatInt(sys_log.RAM_used, 10))
		sb.WriteString("\t")
		sb.WriteString(strconv.FormatFloat(sys_log.DISC, 'f', -1, 64))
		sb.WriteString("\t")
		sb.WriteString(strconv.FormatFloat(sys_log.DISC_used, 'f', -1, 64))
		sb.WriteString("\t")
		sb.WriteString(strconv.Itoa(sys_log.GOMAXPROCS))
		sb.WriteString("\n")
	}
	return []byte(sb.String()), nil
}
