package main

import (
	"flag"
	"github.com/shamank/edutour-backend/auth-service/internal/app"
)

const (
	configDir = "./configs"

	prodConfig  = "prod.yml"
	localConfig = "local.yml"
)

func main() {
	isProd := flag.Bool("prod", false, "enable production mode")

	// Парсинг флагов
	flag.Parse()

	if *isProd {
		app.Run(configDir + "/" + prodConfig)
		return
	}
	app.Run(configDir + "/" + localConfig)
}
