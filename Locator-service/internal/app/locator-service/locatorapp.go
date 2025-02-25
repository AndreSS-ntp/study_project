package locator_service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/domain"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Command struct {
	Description string
	Handler     func(http.ResponseWriter, *http.Request)
}

type App struct {
	Commands map[string]Command
	RWMutex  *sync.RWMutex
}

func NewApp(mu *sync.RWMutex) *App {
	a := App{}
	var commands = map[string]Command{
		"/service-health/{id}": Command{"Состояние сервиса под ID.", a.ServiceHealth},
	}
	a.Commands = commands
	a.RWMutex = mu
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

	file, err_open := os.Open(config.PathLogs)
	if err_open != nil {
		err_open = fmt.Errorf("500 - internal server error: %w", err_open)
		w.WriteHeader(500)
		_, w_err := w.Write([]byte(err_open.Error()))
		if w_err != nil {
			w_err = fmt.Errorf("cant write a response: %w", w_err)
			fmt.Println(w_err)
		}
		return
	}
	defer func() {
		err_close := file.Close()
		if err_close != nil {
			err_close = fmt.Errorf("cant close logFile: %w", err_close)
			fmt.Println(err_close)
		}
	}()

	in := bufio.NewReader(file)

	sb := strings.Builder{}
	for {
		a.RWMutex.RLock()
		line, err_read := in.ReadString('\n')
		a.RWMutex.RUnlock()
		if err_read != nil {
			if err_read == io.EOF {
				break
			} else {
				err_read = fmt.Errorf("500 - internal server error: %w", err_read)
				w.WriteHeader(500)
				_, w_err := w.Write([]byte(err_read.Error()))
				if w_err != nil {
					w_err = fmt.Errorf("cant write a response: %w", w_err)
					fmt.Println(w_err)
				}
				return
			}
		}
		splited_line := strings.Split(line, " ")
		parsed_id, err_parse := strconv.ParseInt(splited_line[0], 10, 64)
		if err_parse != nil {
			err_parse = fmt.Errorf("500 - internal server error: %w", err_parse)
			w.WriteHeader(500)
			_, w_err := w.Write([]byte(err_parse.Error()))
			if w_err != nil {
				w_err = fmt.Errorf("cant write a response: %w", w_err)
				fmt.Println(w_err)
			}
			return
		}

		if parsed_id == id {
			sys_log := domain.System{}
			err_unmarsh := json.Unmarshal([]byte(splited_line[3]), &sys_log)
			if err_unmarsh != nil {
				err_unmarsh = fmt.Errorf("500 - internal server error: %w", err_unmarsh)
				w.WriteHeader(500)
				_, w_err := w.Write([]byte(err_unmarsh.Error()))
				if w_err != nil {
					w_err = fmt.Errorf("cant write a response: %w", w_err)
					fmt.Println(w_err)
				}
				return
			}
			i := -1
			var cpu_name string
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

			bytes_to_grow := len(splited_line[1]) + len(splited_line[2]) + 2 +
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
			sb.WriteString(splited_line[1])
			sb.WriteString("\t")
			sb.WriteString(splited_line[2])
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
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/csv")
	_, w_err := w.Write([]byte(sb.String()))
	if w_err != nil {
		w_err = fmt.Errorf("cant write a response: %w", w_err)
		fmt.Println(w_err)
	}
}
