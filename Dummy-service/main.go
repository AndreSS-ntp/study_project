package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const portNum string = ":7001"

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type System struct {
	num_CPU    int
	CPU_usage  map[string]float64
	RAM        int64
	RAM_used   int64
	DISC       float64
	DISC_used  float64
	GOMAXPROCS int
}

var commands = map[string]string{
	"/help":   "Список команд.",
	"/health": "Вернуть состояние сервиса и данные о системе сервера.",
}

func help(w http.ResponseWriter, r *http.Request) {
	message := ""
	for key, value := range commands {
		message += key + " - " + value + "\n"
	}
	fmt.Fprintf(w, message)
}

func health(w http.ResponseWriter, r *http.Request) {
	var system System
	system.num_CPU = runtime.NumCPU()
	system.CPU_usage = countCPUusage()
	system.RAM, system.RAM_used = getRAMSample()
	system.DISC, system.DISC_used = getDISCSample("/")
	system.GOMAXPROCS = runtime.GOMAXPROCS(0)
	data, error := json.Marshal(system)
	if error != nil {
		fmt.Fprintf(w, "ERROR: marshling ", error)
	}
	fmt.Fprintf(w, string(data))
}

func getRAMSample() (int64, int64) {
	var MemTotal, MemAvailable int64
	contents, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		fmt.Println(err)
		return 0, 0
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines[:len(lines)-1] {
		fields := strings.Fields(line)

		switch fields[0] {
		case "MemTotal:":
			MemTotal, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				fmt.Println("Error: ", fields, err)
			}
		case "MemAvailable:":
			MemAvailable, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				fmt.Println("Error: ", fields, err)
			}
		}
	}
	return MemTotal, MemTotal - MemAvailable

}

func getCPUSample() (map[string]uint64, map[string]uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	total := make(map[string]uint64)
	total_idle := make(map[string]uint64)
	if err != nil {
		fmt.Println(err)
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
					fmt.Println("Error: ", i, fields[i], err)
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

func countCPUusage() map[string]float64 {
	total_0, total_idle_0 := getCPUSample()
	time.Sleep(5 * time.Second)
	total_1, total_idle_1 := getCPUSample()
	CPU_usage := make(map[string]float64)

	idleTicks := make(map[string]float64)
	totalTicks := make(map[string]float64)

	if len(total_0) == 0 || len(total_1) == 0 {
		fmt.Println("ERROR: zero data from <getCPUsample>")
		return CPU_usage
	}

	for key, _ := range total_0 {
		totalTicks[key] = float64(total_1[key] - total_0[key])
		idleTicks[key] = float64(total_idle_1[key] - total_idle_0[key])

		CPU_usage[key] = 100 * (totalTicks[key] - idleTicks[key]) / totalTicks[key]
	}

	return CPU_usage
}

func getDISCSample(path string) (float64, float64) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return 0, 0
	}
	disk_ALL := fs.Blocks * uint64(fs.Bsize)
	disk_FREE := fs.Bfree * uint64(fs.Bsize)
	disk_USED := disk_ALL - disk_FREE

	return float64(disk_ALL) / GB, float64(disk_USED) / GB
}

func main() {
	http.HandleFunc("/help", help)
	http.HandleFunc("/health", health)

	err := http.ListenAndServe(portNum, nil)
	if err != nil {
		log.Fatal(err)
	}
}
