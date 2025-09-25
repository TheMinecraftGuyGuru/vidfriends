package main

import (
	"context"
	"log"
	"os"

	"github.com/vidfriends/backend/internal/app"
)

func main() {
	ctx := context.Background()
	if err := app.Run(ctx, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
