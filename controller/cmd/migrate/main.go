// Command migrate applies or rolls back database migrations.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/managerhub/managerhub/controller/internal/db"
)

func main() {
	url := os.Getenv("MH_DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "MH_DATABASE_URL is required")
		os.Exit(1)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer pool.Close()

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "up":
		err = db.Migrate(ctx, pool)
	case "down":
		err = db.RollbackDown(ctx, pool)
	default:
		err = fmt.Errorf("unknown command %q (use up|down)", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("migrate:", cmd, "ok")
}
