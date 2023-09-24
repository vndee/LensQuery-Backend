package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Pool *gorm.DB

func GetCloudSQLDB() (func() error, error) {
	if Pool != nil {
		return nil, nil
	}

	mustGetenv := func(k string) string {
		v := os.Getenv(k)
		if v == "" {
			log.Fatalf("Fatal Error in connect_unix.go: %s environment variable not set.", k)
		}
		return v
	}

	var (
		dbUser         = mustGetenv("DB_USER")                                    // e.g. 'my-db-user'
		dbPwd          = mustGetenv("DB_PASS")                                    // e.g. 'my-db-password'
		dbName         = mustGetenv("DB_NAME")                                    // e.g. 'my-database'
		unixSocketPath = "/cloudsql/" + mustGetenv("DB_INSTANCE_CONNECTION_NAME") // e.g. '/cloudsql/project:region:instance'
	)

	dbURI := fmt.Sprintf("%s:%s@unix(%s)/%s?parseTime=true",
		dbUser, dbPwd, unixSocketPath, dbName)

	// dbPool is the pool of database connections.
	dbConn, err := sql.Open("postgres", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// cleanup, err := pgxv4.RegisterDriver("cloudsql-postgres", cloudsqlconn.WithIAMAuthN())
	// if err != nil {
	// 	log.Fatalf("Error on pgxv4.RegisterDriver: %v", err)
	// }
	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_INSTANCE_CONNECTION_NAME"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	// db, err := sql.Open("cloudsql-postgres", dsn)
	// if err != nil {
	// 	log.Fatalf("Error on sql.Open: %v", err)
	// }

	// // Ping to make sure the database is accessible and works.
	// if err := db.Ping(); err != nil {
	// 	log.Fatalf("Error on db.Ping: %v", err)
	// }

	dbPool, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbConn,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error on gorm.Open: %v", err)
	}

	sqlDB, err := dbPool.DB()
	if err != nil {
		log.Fatalf("Error on dbPool.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	Pool = dbPool

	// check if the database is accessible
	if err := Pool.Exec("SELECT 1").Error; err != nil {
		log.Fatalf("Error on pool.Exec: %v", err)
	}

	log.Println("Connected to Cloud SQL")
	return nil, nil
}

func CreateTables() {
	Pool.AutoMigrate(&model.UserCredits{})
	Pool.AutoMigrate(&model.CreditUsageHistory{})
}

func ProcessDatabaseResponse(response *gorm.DB) error {
	if response.Error != nil {
		return response.Error
	}

	if response.RowsAffected == 0 {
		return fiber.ErrNotFound
	}

	return nil
}
