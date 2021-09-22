package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"sync"
	"time"

	rd "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/chi3ndd/library/clock"
)

type redisConnector struct {
	con    *rd.Client
	mutex  *sync.Mutex
	config Config
}

func NewService(conf Config, tlsConf *tls.Config) Service {
	con := rd.NewClient(&rd.Options{
		Addr:      conf.Address,
		Password:  conf.Password,
		DB:        conf.Db,
		TLSConfig: tlsConf,
	})
	r := &redisConnector{con: con, mutex: &sync.Mutex{}, config: conf}
	// Monitor
	r.monitor()
	// Success
	return r
}

func (r *redisConnector) monitor() {
	// Reconnect connection
	go func() {
		for {
			if r.con != nil {
				if r.Ping() != nil {
					logrus.Info("[Redis] Connection closed")
					r.mutex.Lock()
					for {
						r.con = rd.NewClient(&rd.Options{
							Addr:      r.config.Address,
							Password:  r.config.Password,
							DB:        r.config.Db,
							TLSConfig: nil,
						})
						if r.Ping() == nil {
							logrus.Info("[Redis] Reconnect success!")
							break
						}
						logrus.Info("[Redis] Reconnecting ...")
						// Sleep
						clock.Sleep(clock.Second * 3)
					}
					r.mutex.Unlock()
				}
			}
			// Sleep
			clock.Sleep(clock.Second * 10)
		}
	}()
}

func (r *redisConnector) Ping() error {
	// Success
	return r.con.Ping(context.Background()).Err()
}

func (r *redisConnector) Delete(keys ...string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// Success
	return r.con.Del(context.Background(), keys...).Err()
}

func (r *redisConnector) Expire(key string, ttl clock.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// Success
	return r.con.Expire(context.Background(), key, time.Duration(ttl)).Err()
}

func (r *redisConnector) ExpireAt(key string, tm time.Time) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// Success
	return r.con.ExpireAt(context.Background(), key, tm).Err()
}

func (r *redisConnector) Set(key, value string, ttl clock.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// Success
	return r.con.Set(context.Background(), key, value, time.Duration(ttl)).Err()
}

func (r *redisConnector) SetObject(key string, value interface{}, ttl clock.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	bts, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// Success
	return r.Set(key, string(bts), ttl)
}

func (r *redisConnector) Get(key string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	// Success
	result, err := r.con.Get(context.Background(), key).Result()
	if err == rd.Nil {
		return result, errors.New(NotFoundError)
	}
	return result, nil
}

func (r *redisConnector) GetObject(key string, pointer interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	result, err := r.Get(key)
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(result), pointer); err != nil {
		return err
	}
	// Success
	return nil
}
