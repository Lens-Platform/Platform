package api

import (
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version"
)

func (s *Server) startCachePool(ticker *time.Ticker, stopCh <-chan struct{}) {
	if s.config.CacheServer == "" {
		return
	}
	s.pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.config.CacheServer)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	// set <hostname>=<version> with an expiry time of one minute
	setVersion := func() {
		conn := s.pool.Get()
		if _, err := conn.Do("SET", s.config.Hostname, version.VERSION, "EX", 60); err != nil {
			s.logger.Error(err, "cache server is offline", "error", err.Error(), "server", s.config.CacheServer)
		}
		_ = conn.Close()
	}

	// set version on a schedule
	go func() {
		setVersion()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				setVersion()
			}
		}
	}()
}
