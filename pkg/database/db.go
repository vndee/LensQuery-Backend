package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"cloud.google.com/go/cloudsqlconn/postgres/pgxv4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Pool *gorm.DB

func GetCloudSQLDB() (func() error, error) {
	if Pool != nil {
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

	// Ping to make sure the database is accessible and works.
	if err := db.Ping(); err != nil {
		log.Fatalf("Error on db.Ping: %v", err)
	}

	dbPool, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error on gorm.Open: %v", err)
	}

	Pool = dbPool

	// check if the database is accessible
	if err := Pool.Exec("SELECT 1").Error; err != nil {
		log.Fatalf("Error on pool.Exec: %v", err)
	}

	log.Println("Connected to Cloud SQL")
	return cleanup, nil
}
