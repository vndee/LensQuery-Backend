package database

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Pool *gorm.DB

func GetCloudSQLDB() error {
	if Pool != nil {
		return nil
	}

	dbConn, err := ConnectGCPWithConnector()
	if err != nil {
		log.Fatalf("Error on ConnectWithConnector: %v", err)
	}

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
	return nil
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
