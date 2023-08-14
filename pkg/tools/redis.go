package tools

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

func NewRedis(url, password string, db int) (*redis.Pool, error) {

	return &redis.Pool{
		MaxIdle:     2,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", url)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if db < 0 || db > 15 {
				c.Close()
				return nil, err
			}

			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}, nil
}
