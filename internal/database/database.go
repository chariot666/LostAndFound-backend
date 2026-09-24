package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"lost-found-server/internal/config"
	"lost-found-server/internal/model"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Item{},
		&model.Claim{},
		&model.Announcement{},
	); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err := db.Exec("ALTER TABLE users AUTO_INCREMENT = 10001").Error; err != nil {
		return nil, fmt.Errorf("set users auto increment: %w", err)
	}

	return db, nil
}
