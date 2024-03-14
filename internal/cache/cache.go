package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/rueidis"

	"github.com/tlipoca9/asta/internal/config"
)

var (
	_s    Service
	_init sync.Once
)

type Service interface {
	Health() bool
}

type service struct {
	log    *slog.Logger
	client rueidis.Client
}

type Config struct {
	Address string
}

func New() Service {
	if _s == nil {
		_init.Do(func() {
			_s = newService(slog.Default(), Config{
				Address: config.C.Cache.Address,
			})
		})
	}
	return _s
}

func newService(log *slog.Logger, conf Config) Service {
	cli, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{conf.Address}})
	if err != nil {
		panic(err)
	}
	config.DeferShutdown("cache", cli.Close)

	s := &service{
		log:    log,
		client: cli,
	}
	return s
}

func (s *service) Health() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	pingCmd := s.client.B().Ping().Build()
	err := s.client.Do(ctx, pingCmd).Error()
	if err != nil {
		s.log.Error("failed to ping cache", "error", err)
		return false
	}

	return true
}
