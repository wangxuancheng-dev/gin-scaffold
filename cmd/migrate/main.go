package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/joho/godotenv"
	cli "github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	app := &cli.App{
		Name:  "migrate",
		Usage: "database migration tool",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "env", Value: "dev", Usage: "runtime env, dev will auto load .env files"},
			&cli.StringFlag{Name: "profile", Usage: "optional profile used by .env.<env>.<profile>"},
			&cli.StringFlag{Name: "driver", Value: "mysql", Usage: "mysql|postgres"},
			&cli.StringFlag{Name: "dsn", Usage: "database dsn, fallback to DB_DSN env"},
			&cli.StringFlag{Name: "dir", Value: "./migrations", Usage: "migration directory"},
		},
		Commands: []*cli.Command{
			{
				Name: "up",
				Action: func(c *cli.Context) error {
					loadDotEnv(c.String("env"), c.String("profile"))
					db, err := openDB(c.String("driver"), resolveDSN(c.String("dsn")))
					if err != nil {
						return err
					}
					m := buildMigrator(db, c.String("dir"))
					if err = m.Migrate(); err != nil {
						return err
					}
					fmt.Println("migrate up done")
					return nil
				},
			},
			{
				Name: "down",
				Action: func(c *cli.Context) error {
					loadDotEnv(c.String("env"), c.String("profile"))
					db, err := openDB(c.String("driver"), resolveDSN(c.String("dsn")))
					if err != nil {
						return err
					}
					m := buildMigrator(db, c.String("dir"))
					if err = m.RollbackLast(); err != nil {
						return err
					}
					fmt.Println("migrate down one step done")
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func resolveDSN(flagDSN string) string {
	if flagDSN != "" {
		return flagDSN
	}
	return os.Getenv("DB_DSN")
}

func loadDotEnv(env, profile string) {
	if env != "dev" {
		return
	}
	files := []string{
		".env",
		fmt.Sprintf(".env.%s", env),
	}
	if profile != "" {
		files = append(files, fmt.Sprintf(".env.%s.%s", env, profile))
	}
	files = append(files, ".env.local", fmt.Sprintf(".env.%s.local", env))
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			_ = godotenv.Load(f)
		}
	}
}

func openDB(driver, dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty dsn: use --dsn or set DB_DSN")
	}
	var dialector gorm.Dialector
	switch driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func buildMigrator(db *gorm.DB, dir string) *gormigrate.Gormigrate {
	upPath := filepath.Join(dir, "20250101_init.up.sql")
	downPath := filepath.Join(dir, "20250101_init.down.sql")
	migrations := []*gormigrate.Migration{
		{
			ID: "20250101_init",
			Migrate: func(tx *gorm.DB) error {
				sqlBytes, err := os.ReadFile(upPath)
				if err != nil {
					return err
				}
				return tx.Exec(string(sqlBytes)).Error
			},
			Rollback: func(tx *gorm.DB) error {
				sqlBytes, err := os.ReadFile(downPath)
				if err != nil {
					return err
				}
				return tx.Exec(string(sqlBytes)).Error
			},
		},
	}
	return gormigrate.New(db, gormigrate.DefaultOptions, migrations)
}
