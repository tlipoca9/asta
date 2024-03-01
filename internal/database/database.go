package database

import (
	"context"
	"fmt"
	"time"

	"github.com/tlipoca9/asta/pkg/logx"
	"github.com/tlipoca9/errors"
	"github.com/tlipoca9/leaf/gormleaf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Service interface {
	Health() error
}

type service struct {
	db *gorm.DB
}

type Config struct {
	DBName   string
	Password string
	Username string
	Port     string
	Host     string
}

func New(conf Config) Service {
	log := logx.New()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Host, conf.Port, conf.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               gormleaf.NewSlogLoggerBuilder().Logger(log).Build(),
	})
	if err != nil {
		log.Error("failed to connect to database", "error", err, "dsn", dsn)
		panic(err)
	}

	s := &service{db: db}
	return s
}

func (s *service) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	db, err := s.db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get db")
	}

	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "db down")
	}

	return nil
}
