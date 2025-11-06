package config

import "os"

const IP_port string = "0.0.0.0:7002"

var DB_URL = os.Getenv("DB_URL")
