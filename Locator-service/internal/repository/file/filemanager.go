package file

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	alogger "github.com/AndreSS-ntp/logger"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/config"
	"github.com/unwisecode/over-the-horison-andress/tree/main/Locator-service/internal/domain"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

var mutex = &sync.RWMutex{}

type FileManager struct {
	logPath string
}

func NewFileManager(LogsFilePath string) *FileManager {
	return &FileManager{LogsFilePath}
}

func (fm *FileManager) createLogFile(path string) (*os.File, error) {
	err_make := os.MkdirAll(config.PathLogs, os.ModePerm)
	if err_make != nil {
		err_make = fmt.Errorf("cant make dir with logs: %w", err_make)
		return nil, err_make
	}
	logFile, err := os.OpenFile(config.PathLogs, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = fmt.Errorf("cant open logFile: %w", err)
		return nil, err
	}
	return logFile, nil
}

func (fm *FileManager) WriteLog(ctx context.Context, log string) error {
	mutex.Lock()
	defer mutex.Unlock()
	logFile, err := os.OpenFile(config.PathLogs, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if errors.Is(err, os.ErrNotExist) {
		logFile, err = fm.createLogFile(log)
	}
	if err != nil {
		err = fmt.Errorf("cant open logFile: %w", err)
		return err
	}
	defer func() {
		err_close := logFile.Close()
		if err_close != nil {
			alogger.FromContext(ctx).Error(ctx, "cant close logFile: "+err_close.Error())
		}
	}()

	_, err_w := logFile.Write([]byte(log))
	if err_w != nil {
		err = fmt.Errorf("write error occured: %w", err)
		return err
	}
	return nil
}

func (fm *FileManager) GetLogsById(ctx context.Context, id int64) ([]domain.System, [][]string, error) {
	logs := make([]domain.System, 0)
	timestamps := make([][]string, 0)

	mutex.RLock()
	defer mutex.RUnlock()

	file, err_open := os.Open(fm.logPath)
	if err_open != nil {
		err_open = fmt.Errorf("500 - internal server error: %w", err_open)
		return nil, nil, err_open
	}
	defer func() {
		err_close := file.Close()
		if err_close != nil {
			alogger.FromContext(ctx).Error(ctx, "cant close logFile: "+err_close.Error())
		}
	}()

	in := bufio.NewReader(file)

	for {
		line, err_read := in.ReadString('\n')
		if err_read != nil {
			if err_read == io.EOF {
				break
			} else {
				err_read = fmt.Errorf("500 - internal server error: %w", err_read)
				return nil, nil, err_read
			}
		}
		splited_line := strings.Split(line, " ")
		parsed_id, err_parse := strconv.ParseInt(splited_line[0], 10, 64)
		if err_parse != nil {
			err_parse = fmt.Errorf("500 - internal server error: %w", err_parse)
			return nil, nil, err_parse
		}

		if parsed_id == id {
			sys_log := domain.System{}
			err_unmarsh := json.Unmarshal([]byte(splited_line[3]), &sys_log)
			if err_unmarsh != nil {
				err_unmarsh = fmt.Errorf("500 - internal server error: %w", err_unmarsh)
				return nil, nil, err_unmarsh
			}
			logs = append(logs, sys_log)
			timestamps = append(timestamps, []string{splited_line[1], splited_line[2]})
		}
	}
	return logs, timestamps, nil
}
