package dlock

import (
	"github.com/choral-io/gommerce-server-core/config"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidislock"
)

type Locker = rueidislock.Locker

// NewRedisLocker creates a new rueidislock.Locker with the given config.
func NewRedisLocker(rdb rueidis.Client, cfg config.ServerLockerConfig) (rueidislock.Locker, error) {
	return rueidislock.NewLocker(rueidislock.LockerOption{
		ClientBuilder: func(opts rueidis.ClientOption) (rueidis.Client, error) {
			return rdb, nil
		},
		KeyPrefix:      cfg.GetKeyPrefix(),
		KeyValidity:    cfg.GetKeyValidity(),
		ExtendInterval: cfg.GetExtendInterval(),
		TryNextAfter:   cfg.GetTryNextAfter(),
		KeyMajority:    cfg.GetKeyMajority(),
		NoLoopTracking: cfg.GetNoLoopTracking(),
		FallbackSETPX:  cfg.GetFallbackSETPX(),
	})
}
