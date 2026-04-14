package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(fmt.Errorf("read migration dir: %w", err))
	}
	upFiles := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			upFiles = append(upFiles, name)
		}
	}
	sort.Strings(upFiles)

	migrations := make([]*gormigrate.Migration, 0, len(upFiles))
	for _, upName := range upFiles {
		base := strings.TrimSuffix(upName, ".up.sql")
		upPath := filepath.Join(dir, upName)
		downPath := filepath.Join(dir, base+".down.sql")
		mID := base
		migrations = append(migrations, &gormigrate.Migration{
			ID: mID,
			Migrate: func(tx *gorm.DB) error {
				sqlBytes, err := os.ReadFile(upPath)
				if err != nil {
					return err
				}
				return execSQLStatements(tx, string(sqlBytes))
			},
			Rollback: func(tx *gorm.DB) error {
				if _, err := os.Stat(downPath); err != nil {
					return nil
				}
				sqlBytes, err := os.ReadFile(downPath)
				if err != nil {
					return err
				}
				return execSQLStatements(tx, string(sqlBytes))
			},
		})
	}
	return gormigrate.New(db, gormigrate.DefaultOptions, migrations)
}

func execSQLStatements(tx *gorm.DB, script string) error {
	for _, stmt := range splitSQLStatements(script) {
		if err := tx.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func splitSQLStatements(script string) []string {
	var (
		out      []string
		buf      strings.Builder
		inSingle bool
		inDouble bool
		prev     rune
	)
	for _, r := range script {
		switch r {
		case '\'':
			if !inDouble && prev != '\\' {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle && prev != '\\' {
				inDouble = !inDouble
			}
		case ';':
			if !inSingle && !inDouble {
				s := strings.TrimSpace(buf.String())
				if s != "" {
					out = append(out, s)
				}
				buf.Reset()
				prev = r
				continue
			}
		}
		buf.WriteRune(r)
		prev = r
	}
	if s := strings.TrimSpace(buf.String()); s != "" {
		out = append(out, s)
	}
	return out
}
