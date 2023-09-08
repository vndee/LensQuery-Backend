package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"cloud.google.com/go/cloudsqlconn/postgres/pgxv4"
	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBInstance struct {
	DB *gorm.DB
}

var DB *DBInstance
var database *sql.DB

func Connect() error {
	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Missing required environment variable %s", k)
		}
		return v
	}

	var (
		dbUser = mustGetenv("DB_USER")
		dbPwd  = mustGetenv("DB_PASS")
		dbHost = mustGetenv("DB_HOST")
		dbPort = mustGetenv("DB_PORT")
		dbName = mustGetenv("DB_NAME")
	)

	dbURI := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", dbHost, dbUser, dbPwd, dbName, dbPort)

	dbPool, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("sql.Open: %v", err)
	}
	dbPool.AutoMigrate(&model.UserCredits{}, &model.UserSubscription{}, &model.SubcriptionPlan{}, &model.CreditUsageHistory{})

	DB = &DBInstance{
		DB: dbPool,
	}
	return nil
}

func GetCloudSQLDB() (func() error, error) {
	if database != nil {
		return nil, nil
	}
	cleanup, err := pgxv4.RegisterDriver("cloudsql-postgres", cloudsqlconn.WithIAMAuthN())
	if err != nil {
		log.Fatalf("Error on pgxv4.RegisterDriver: %v", err)
	}
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable", os.Getenv("DB_INSTANCE_CONNECTION_NAME"), os.Getenv("DB_USER"), os.Getenv("DB_NAME"))
	db, err := sql.Open("cloudsql-postgres", dsn)
	if err != nil {
		log.Fatalf("Error on sql.Open: %v", err)
	}

	database = db

	// Ping to make sure the database is accessible and works.
	if err := db.Ping(); err != nil {
		log.Fatalf("Error on db.Ping: %v", err)
	}

	log.Println("Connected to Cloud SQL")
	return cleanup, nil
}
