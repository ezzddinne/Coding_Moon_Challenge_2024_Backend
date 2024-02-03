package server

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/ezzddinne/CMC_BACKEND/api"
	"github.com/ezzddinne/CMC_BACKEND/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("[WARNING]", err)
	}
}

// database connection
func DBConnection() (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			Colorful:      true,
			LogLevel:      logger.Info,
		},
	)

	url := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	return gorm.Open(postgres.Open(url), &gorm.Config{Logger: newLogger})
}

// run database
func RunServer() {
	db, err := DBConnection()
	if err != nil {
		panic(fmt.Sprintf("[WARNING] database connection: %v", err))
	}

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		panic(fmt.Sprintf("[WARNING] failed to initialize casbin adapter: %v", err))
	}

	enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
	if err != nil {
		panic(fmt.Sprintf("[WARNING] failed to create casbin enforcer: %v", err))
	}

	database_flag := flag.Bool("database", false, "Bool variable to create database")
	flag.Parse()

	if *database_flag {
		database.AutoMigrateDatabase(db, enforcer)
		return
	}

	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"}, // Only allow http://localhost:4200
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"}, // Include Authorization
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router_api := router.Group("/api")
	{

		api.RoutesApis(router_api, db, enforcer)
	}

	router.Run(os.Getenv("APP_PORT"))
}
