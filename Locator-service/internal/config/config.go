package config

const IP_port string = "0.0.0.0:7000"

var Adresses = map[int64]string{
	1: "http://172.17.0.1:7001",
	// 1: "http://127.0.0.1:7001", // для тестов без докера
}

const PathLogs string = "./internal/repository/file-storage/system-data/logs"
