package main

import (
	"flag"
	"github.com/shamank/edutour-backend/apigateway/internal/app"
)

const (
	configDir = "./configs"
)

func main() {

	prodFlag := flag.Bool("prod", false, "start on production")
	//localFlag := flag.Bool("local", false, "start on local")

	flag.Parse()

	var resultCfg string

	if *prodFlag {
		resultCfg = configDir + "/prod.yaml"

	} else {
		resultCfg = configDir + "/local.yaml"
	}

	app.Run(resultCfg)
}
