package config

import "os"

const IP_port string = "0.0.0.0:7003"

var DB_URL = os.Getenv("STORAGE_DB_URL")
