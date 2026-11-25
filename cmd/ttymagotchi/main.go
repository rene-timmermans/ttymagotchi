package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/rene-timmermans/ttymagotchi/internal/app"
	"github.com/rene-timmermans/ttymagotchi/internal/config"
)

var version = "0.3.1-dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("ttymagotchi " + version)
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	a := app.NewApp(cfg)
	a.Run()
}
