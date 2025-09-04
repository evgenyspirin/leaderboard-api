package main

import (
	"context"
	"log"
	"os"

	"leaderboard-api/internal"
)

// I focused on implementing more important features and left this list for later.
// todo: Implement:
// todo: Panic recovery
// todo: Validation all requests
// todo: Linters(GolangCiLint)
// todo: Authorization for endpoints(in case of external requests)
// todo: Central error handling pattern "SPE":
// https://medium.com/@yevheniikulhaviuk/golang-architectural-pattern-for-errors-531c0e54d67b

func main() {
	ctx := context.Background()

	app, err := internal.NewApp(ctx)
	if err != nil {
		log.Fatalf("init app failed: %v", err)
	}
	defer app.Close()

	app.InitControllers(ctx)

	if err = app.Run(ctx); err != nil {
		app.Logger().Sugar().Errorf("leaderboardapi stopped with error: %v", err)
		os.Exit(1)
	}
}
