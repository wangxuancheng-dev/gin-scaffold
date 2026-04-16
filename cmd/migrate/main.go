package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	scaffolddb "gin-scaffold/pkg/db"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var env string
	var profile string
	var driver string
	var dsn string
	var timeZone string
	var dir string

	rootCmd := &cobra.Command{
		Use:   "migrate",
		Short: "database migration tool",
	}
	rootCmd.PersistentFlags().StringVar(&env, "env", "dev", "runtime env, dev will auto load .env files")
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "optional profile used by .env.<env>.<profile>")
	rootCmd.PersistentFlags().StringVar(&driver, "driver", "mysql", "mysql|postgres")
	rootCmd.PersistentFlags().StringVar(&dsn, "dsn", "", "database dsn, fallback to DB_DSN env")
	rootCmd.PersistentFlags().StringVar(&timeZone, "time-zone", "", "mysql SET time_zone / pg SET TIME ZONE; env TIME_ZONE; default UTC")
	rootCmd.PersistentFlags().StringVar(&dir, "dir", "", "migration directory, default auto: ./migrations/<driver>")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "apply all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			loadDotEnv(env, profile)
			driverName := normalizeDriver(driver)
			db, err := openDB(driverName, resolveDSN(dsn), resolveTimeZone(timeZone))
			if err != nil {
				return err
			}
			migrationDir, err := resolveMigrationDir(dir, driverName)
			if err != nil {
				return err
			}
			m := buildMigrator(db, migrationDir)
			if err = m.Migrate(); err != nil {
				return err
			}
			fmt.Println("migrate up done")
			return nil
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "rollback one migration step",
		RunE: func(cmd *cobra.Command, args []string) error {
			loadDotEnv(env, profile)
			driverName := normalizeDriver(driver)
			db, err := openDB(driverName, resolveDSN(dsn), resolveTimeZone(timeZone))
			if err != nil {
				return err
			}
			migrationDir, err := resolveMigrationDir(dir, driverName)
			if err != nil {
				return err
			}
			m := buildMigrator(db, migrationDir)
			if err = m.RollbackLast(); err != nil {
				return err
			}
			fmt.Println("migrate down one step done")
			return nil
		},
	})

	if err := rootCmd.Execute(); err != nil {
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

func resolveTimeZone(cliVal string) string {
	if strings.TrimSpace(cliVal) != "" {
		return strings.TrimSpace(cliVal)
	}
	if v := strings.TrimSpace(os.Getenv("TIME_ZONE")); v != "" {
		return v
	}
	return "UTC"
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "pg":
		return "postgres"
	case "postgres":
		return "postgres"
	default:
		return "mysql"
	}
}

func resolveMigrationDir(dirFlag, driver string) (string, error) {
	if strings.TrimSpace(dirFlag) != "" {
		return dirFlag, nil
	}
	candidate := filepath.Join("migrations", driver)
	if st, err := os.Stat(candidate); err == nil && st.IsDir() {
		return candidate, nil
	}
	// Backward compatibility: existing projects may still store mysql sql in ./migrations.
	if driver == "mysql" {
		if st, err := os.Stat("migrations"); err == nil && st.IsDir() {
			return "migrations", nil
		}
	}
	return "", fmt.Errorf("migration dir not found for driver=%s, expected: %s", driver, candidate)
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

func openDB(driver, dsn, sessionTZ string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty dsn: use --dsn or set DB_DSN")
	}
	tzNorm := scaffolddb.NormalizeTimeZone(sessionTZ)
	loc, err := scaffolddb.LocationForTimeZone(tzNorm)
	if err != nil {
		return nil, err
	}

	var dialector gorm.Dialector
	switch driver {
	case "mysql":
		dsn, err = scaffolddb.NormalizeMySQLDSN(dsn, loc)
		if err != nil {
			return nil, err
		}
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
	if err = scaffolddb.ApplyTimeZone(db, driver, sessionTZ); err != nil {
		return nil, fmt.Errorf("set time zone: %w", err)
	}
	return db, nil
}

func buildMigrator(db *gorm.DB, dir string) *gormigrate.Gormigrate {
	upFiles := make([]string, 0, 16)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".up.sql") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		upFiles = append(upFiles, rel)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("scan migration dir: %w", err))
	}
	sort.Strings(upFiles)

	migrations := make([]*gormigrate.Migration, 0, len(upFiles))
	for _, upRel := range upFiles {
		baseRel := strings.TrimSuffix(upRel, ".up.sql")
		upPath := filepath.Join(dir, upRel)
		downPath := filepath.Join(dir, baseRel+".down.sql")
		mID := filepath.Base(baseRel)
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
