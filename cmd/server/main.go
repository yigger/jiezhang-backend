package main

import (
	"log"

	"github.com/yigger/jiezhang-backend/internal/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	if err := app.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
