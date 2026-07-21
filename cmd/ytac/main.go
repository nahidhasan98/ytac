package main

import (
	"os"

	"github.com/nahidhasan98/ytac/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
