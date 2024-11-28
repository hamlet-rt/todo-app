package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo"
	"todo/pkg/handler"
	"todo/pkg/repository"
	"todo/pkg/service"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// @title Todo App API
// @version 1.0
// @description API Server for TodoList Application

// @host localhost:8000
// @basePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializaing configs: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}

	db, err := repository.NewPosgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	scheduler := initScheduler(services)
	scheduler.StartAsync()

	srv := new(todo.Server)
	go func () {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("error ocured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("TodoApp Started")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<- quit

	logrus.Print("TodoApp shutting down")

	scheduler.Stop()

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func initScheduler(services *service.Service) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)

	_, err := scheduler.Every(60).Minutes().Do(func() {
		if err := services.Authorization.DeleteExpiredRefreshTokens(); err != nil {
			logrus.Errorf("Error deleting expired tasks: %v", err)
		} else {
			logrus.Info("Expired tasks deleted successfully")
		}
	})
	if err != nil {
		logrus.Fatalf("Error scheduling task: %s", err.Error())
	}
	return scheduler
}
