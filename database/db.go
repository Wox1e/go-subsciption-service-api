package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"subscription_service/models"
)

type DB struct {
	*gorm.DB
	logger *zap.Logger
}

func Connect(dsn string, logger *zap.Logger) (*DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к БД: %w", err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, fmt.Errorf("Ошибка получения sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("Ошибка ping БД: %w", err)
	}

	logger.Info("Подключение к базе данных установлено")

	return &DB{DB: db, logger: logger}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) CreateSubscription(serviceName string, price int, userID uuid.UUID, startDate, endDate *time.Time) (*uuid.UUID, error) {

	subscription := models.Subscription{
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   *startDate,
		EndDate:     endDate,
	}

	if err := db.Create(&subscription).Error; err != nil {
		db.logger.Error("Ошибка создания подписки",
			zap.String("service_name", serviceName),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("Ошибка создания подписки: %w", err)
	}

	db.logger.Info("Подписка создана",
		zap.String("id", subscription.ID.String()),
		zap.String("service_name", serviceName),
	)

	return &subscription.ID, nil
}

func (db *DB) GetSubscription(id uuid.UUID) (*models.Subscription, error) {
	var subscription models.Subscription
	result := db.Where("id = ?", id).First(&subscription)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			db.logger.Warn("Подписка не найдена", zap.String("id", id.String()))
			return nil, fmt.Errorf("Подписка не найдена")
		}
		db.logger.Error("Ошибка получения подписки",
			zap.String("id", id.String()),
			zap.Error(result.Error),
		)
		return nil, fmt.Errorf("Ошибка получения подписки: %w", result.Error)
	}

	return &subscription, nil
}

func (db *DB) UpdateSubscription(id uuid.UUID, serviceName *string, price *int, startDate, endDate *time.Time) error {
	updates := make(map[string]interface{})

	if serviceName != nil {
		updates["service_name"] = *serviceName
	}
	if price != nil {
		updates["price"] = *price
	}
	if startDate != nil {
		updates["start_date"] = *startDate
	}
	if endDate != nil {
		updates["end_date"] = endDate
	}

	updates["updated_at"] = time.Now()

	result := db.Model(&models.Subscription{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		db.logger.Error("Ошибка обновления подписки",
			zap.String("id", id.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("Ошибка обновления подписки: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		db.logger.Warn("Подписка не найдена для обновления", zap.String("id", id.String()))
		return fmt.Errorf("Подписка не найдена")
	}

	db.logger.Info("Подписка обновлена", zap.String("id", id.String()))
	return nil
}

func (db *DB) DeleteSubscription(id uuid.UUID) error {
	result := db.Where("id = ?", id).Delete(&models.Subscription{})

	if result.Error != nil {
		db.logger.Error("Ошибка удаления подписки",
			zap.String("id", id.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("Ошибка удаления подписки: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		db.logger.Warn("Подписка не найдена для удаления", zap.String("id", id.String()))
		return fmt.Errorf("Подписка не найдена")
	}

	db.logger.Info("Подписка удалена", zap.String("id", id.String()))
	return nil
}

func (db *DB) ListSubscriptions(limit, offset int) ([]models.Subscription, error) {
	var subscriptions []models.Subscription

	result := db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&subscriptions)

	if result.Error != nil {
		db.logger.Error("Ошибка получения списка подписок", zap.Error(result.Error))
		return nil, fmt.Errorf("ошибка получения списка подписок: %w", result.Error)
	}

	db.logger.Info("Список подписок получен", zap.Int("count", len(subscriptions)))
	return subscriptions, nil
}


func (db *DB) CalculateTotalCost(userID *uuid.UUID, serviceName *string, startPeriod, endPeriod time.Time) (int, error) {
	query := db.Model(&models.Subscription{}).
		Where("start_date <= ? AND (end_date IS NULL OR end_date >= ?)", endPeriod, startPeriod)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if serviceName != nil {
		query = query.Where("service_name = ?", *serviceName)
	}

	var totalCost int64
	if err := query.Select("COALESCE(SUM(price), 0)").Scan(&totalCost).Error; err != nil {
		db.logger.Error("Ошибка вычисления суммарной стоимости",
			zap.Error(err),
		)
		return 0, fmt.Errorf("Ошибка вычисления суммарной стоимости: %w", err)
	}

	db.logger.Info("Суммарная стоимость вычислена",
		zap.Int64("total_cost", totalCost),
		zap.Time("start_period", startPeriod),
		zap.Time("end_period", endPeriod),
	)

	return int(totalCost), nil
}
