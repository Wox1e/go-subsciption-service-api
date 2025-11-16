package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB, migrationsDir string, logger *zap.Logger) error {
	logger.Info("Начало выполнения миграций", zap.String("dir", migrationsDir))

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("ошибка получения sql.DB: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка создания драйвера миграций: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("ошибка создания экземпляра migrate: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("Нет новых миграций для применения")
			return nil
		}
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("ошибка получения версии миграции: %w", err)
	}

	if err == nil {
		logger.Info("Миграции выполнены успешно",
			zap.Uint("version", version),
			zap.Bool("dirty", dirty),
		)
	} else {
		logger.Info("Миграции выполнены успешно (нет примененных миграций)")
	}

	return nil
}
