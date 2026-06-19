package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"ariga.io/atlas-go-sdk/atlasexec"
)

func main() {
	workdir, err := atlasexec.NewWorkingDir(
		atlasexec.WithMigrations(os.DirFS("./migrations/versioned")),
	)
	if err != nil {
		log.Fatalf("failed to load working dir: %v", err)
	}
	defer workdir.Close()

	client, err := atlasexec.NewClient(workdir.Path(), "atlas")
	if err != nil {
		log.Fatalf("failed to init atlas client: %v", err)
	}

	res, err := client.MigrateApply(context.Background(), &atlasexec.MigrateApplyParams{
		URL: os.Getenv("DB_DSN"),
	})
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Printf("applied %d migrations\n", len(res.Applied))
	for _, m := range res.Applied {
		fmt.Printf("  ✓ %s\n", m.Name)
	}
}
