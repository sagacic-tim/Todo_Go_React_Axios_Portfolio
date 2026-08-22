// cmd/server/main.go
package main

import (
	"database/sql"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"todo_axios_api/backend/internal/db"
	"todo_axios_api/backend/internal/models"
	"todo_axios_api/backend/internal/transport"
)

func main() {
	cwd, _ := os.Getwd()
	log.Println("🔍 working dir is", cwd)

	if err := godotenv.Load("./.env"); err != nil {
		log.Printf("⚠️ failed to load .env in %s: %v\n", cwd, err)
	}

	// 1) Read and validate env
	user := mustGetenv("DB_USER")
	pass := mustGetenv("DB_PASSWORD")
	host := mustGetenv("DB_HOST")
	port := mustGetenv("DB_PORT")
	name := mustGetenv("DB_NAME")
	mustGetenv("JWT_SECRET") // fail fast if the signing key is absent

	// 2) DSN
	dbURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   net.JoinHostPort(host, port),
		Path:   name,
	}

	q := dbURL.Query()
	q.Set("sslmode", "disable")
	dbURL.RawQuery = q.Encode()

	dsnRaw := dbURL.String()

	// 3) Run migrations with database/sql
	migrateDB, err := sql.Open("postgres", dsnRaw)
	if err != nil {
		log.Fatalf("❌ migrate: open sql: %v", err)
	}
	defer migrateDB.Close()

	if err := migrateDB.Ping(); err != nil {
		log.Fatalf("❌ migrate: ping db: %v", err)
	}

	if err := db.RunMigrations(migrateDB); err != nil {
		log.Fatalf("❌ migrate up failed: %v", err)
	}
	log.Println("✅ migrations applied (or none were pending)")

	// 4) Open GORM
	gormDB, err := gorm.Open(postgres.Open(dsnRaw), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ gorm: open: %v", err)
	}

	// Configure the *runtime* pool used by GORM/handlers
	pool, err := gormDB.DB()
	if err != nil {
		log.Fatalf("❌ gorm: db(): %v", err)
	}

	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(25)
	pool.SetConnMaxLifetime(30 * time.Minute)
	pool.SetConnMaxIdleTime(5 * time.Minute)

	if err := pool.Ping(); err != nil {
		log.Fatalf("❌ gorm: ping db: %v", err)
	}

	// 5) Background: purge expired sessions every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			result := gormDB.
				Where("expires_at < ?", time.Now()).
				Delete(&models.Session{})
			if result.Error != nil {
				log.Printf("⚠️  session cleanup error: %v", result.Error)
			} else if result.RowsAffected > 0 {
				log.Printf("🧹 purged %d expired session(s)", result.RowsAffected)
			}
		}
	}()

	// 6) HTTP server
	router := transport.NewRouter(gormDB)

	addr := ":8080"

	// Log only after we know the address is bindable (avoids "listening" lie).
	if err := assertPortFree(addr); err != nil {
		log.Fatalf("❌ server cannot bind %s: %v", addr, err)
	}

	log.Printf("🚀 server starting on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ server error: %v", err)
	}
}

func mustGetenv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		log.Fatalf("❌ missing required env var: %s", key)
	}
	return v
}

func assertPortFree(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	_ = ln.Close()
	return nil
}
