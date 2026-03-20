package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/yourusername/clawhost/hosting-service/models"
	hostingapi "github.com/yourusername/clawhost/hosting-service/api"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	port := flag.String("port", "8090", "Port for hosting-service API")
	flag.Parse()

	if envPort := os.Getenv("HOSTING_PORT"); envPort != "" {
		*port = envPort
	}

	db, err := openDB()
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(
		&models.Customer{},
		&models.Instance{},
		&models.SupportTicket{},
		&models.SupportMessage{},
		&models.Subscription{},
	); err != nil {
		panic(err)
	}

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	jwtSecret := os.Getenv("HOSTING_JWT_SECRET")
	coreAPIURL := os.Getenv("CORE_API_URL")
	allowDevAuth := strings.EqualFold(os.Getenv("ALLOW_DEV_AUTH"), "true") || strings.EqualFold(os.Getenv("ENVIRONMENT"), "development")

	server := hostingapi.NewServer(db, slog.Default(), stripeKey, stripeWebhookSecret, jwtSecret, coreAPIURL, allowDevAuth)
	if err := server.Start(*port); err != nil {
		panic(err)
	}
}

func openDB() (*gorm.DB, error) {
	if dsn := strings.TrimSpace(os.Getenv("HOSTING_DATABASE_URL")); dsn != "" {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	path := strings.TrimSpace(os.Getenv("HOSTING_SQLITE_PATH"))
	if path == "" {
		path = "hosting-service.db"
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	return db, nil
}
