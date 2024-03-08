package database

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tlipoca9/asta/internal/config"

	"github.com/tlipoca9/leaf/gormleaf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_s    Service
	_init sync.Once
)

type Service interface {
	Health() bool
}

type service struct {
	log *slog.Logger
	db  *gorm.DB
}

type Config struct {
	DBName   string
	Username string
	Password string
	Host     string
	Port     int
}

func New() Service {
	if _s == nil {
		_init.Do(func() {
			_s = newService(slog.Default(), Config{
				DBName:   config.C.Database.DBName,
				Username: config.C.Database.Username,
				Password: config.C.Database.Password,
				Host:     config.C.Database.Host,
				Port:     config.C.Database.Port,
			})
		})
	}
	return _s
}

func newService(log *slog.Logger, conf Config) Service {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Host, conf.Port, conf.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormleaf.NewSlogLoggerBuilder().Logger(log).Build(),
	})
	if err != nil {
		panic(err)
	}

	s := &service{
		log: log,
		db:  db,
	}
	return s
}

func (s *service) Health() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	db, err := s.db.DB()
	if err != nil {
		s.log.Error("failed to get db", "error", err)
		return false
	}

	err = db.PingContext(ctx)
	if err != nil {
		s.log.Error("failed to ping db", "error", err)
		return false
	}

	return true
}
