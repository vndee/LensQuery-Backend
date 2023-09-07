package database

import (
	"fmt"
	"log"
	"os"

	"github.com/vndee/lensquery-backend/pkg/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBInstance struct {
	DB *gorm.DB
}

var DB *DBInstance

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

	dbURI := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh", dbHost, dbUser, dbPwd, dbName, dbPort)

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
