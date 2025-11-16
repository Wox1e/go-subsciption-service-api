package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"subscription_service/config"
	"subscription_service/database"
	"subscription_service/handlers"
)

// @title Subscription Service API
// @version 1.0
// @description REST-сервис для агрегации данных об онлайн-подписках пользователей

// @host localhost:8080
// @BasePath /api/v1
func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("ошибка инициализации логгера: %v", err))
	}
	defer logger.Sync()

	logger.Info("Запуск сервиса подписок")

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Fatal("Ошибка загрузки конфигурации", zap.Error(err))
	}

	db, err := database.Connect(cfg.GetGormDSN(), logger)
	if err != nil {
		logger.Fatal("Ошибка подключения к базе данных", zap.Error(err))
	}
	defer db.Close()

	if err := database.RunMigrations(db.DB, "./migrations", logger); err != nil {
		logger.Fatal("Ошибка выполнения миграций", zap.Error(err))
	}

	router := setupRouter(db, logger)

	logger.Info("Сервер запущен", zap.String("address", cfg.GetServerAddr()))
	if err := http.ListenAndServe(cfg.GetServerAddr(), router); err != nil {
		logger.Fatal("Ошибка запуска сервера", zap.Error(err))
	}
}

func setupRouter(db *database.DB, logger *zap.Logger) *gin.Engine {
	router := gin.Default()

	router.Use(ginLogger(logger))

	subscriptionHandler := handlers.NewSubscriptionHandler(db, logger)

	v1 := router.Group("/api/v1")
	{
		subscriptions := v1.Group("/subscriptions")
		{
			subscriptions.POST("", subscriptionHandler.CreateSubscription)
			subscriptions.GET("", subscriptionHandler.ListSubscriptions)
			subscriptions.GET("/:id", subscriptionHandler.GetSubscription)
			subscriptions.PUT("/:id", subscriptionHandler.UpdateSubscription)
			subscriptions.DELETE("/:id", subscriptionHandler.DeleteSubscription)
			subscriptions.GET("/total-cost", subscriptionHandler.CalculateTotalCost)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("HTTP запрос",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
	}
}

