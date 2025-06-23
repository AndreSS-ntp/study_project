package config

import "time"

const IP_port string = "0.0.0.0:7000"

var Adresses = map[int]string{
	1: "http://172.17.0.1:7001", // dummy
	2: "http://172.17.0.1:7002", // store
	// 1: "http://127.0.0.1:7001", // для тестов без докера
	// 2: "http://127.0.0.1:7002", // для тестов без докера
}

const PathLogs string = "./internal/repository/file-storage/system-data/logs"

const HTTPClientTimeout time.Duration = time.Second * 10
