package store

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"sync"
)

// Const
const (
	RedisStatusOkay              = "OK"
	DefaultRedisConnectionString = "redis://localhost:6379"
	DefaultTTL                   = 0
)

var (
	ErrWithOperation    = "operation completed with error: %s\n"
	OperationSuccessful = "operation successful"
)

// Redis holds the config options and the state of the redis connection through the lifetime of the connection.
type Redis struct {
	connectionString, defaultDB string
	conn                        *redis.Client
	rOpts                       *redis.Options
	logger                      *vlog.Logger
	sync.Mutex
}

// Options implements funcitonal options for passing in Redis configuration options
type Options func(*Redis) error

// NewRedis creates a new instance of Redis, for managing CRUD operations to the configured redis client
func NewRedis(connectionString string, logger *vlog.Logger) (*Redis, error) {
	log := logger.Logger()
	rOpts, err := redis.ParseURL(connectionString)
	if err != nil {
		log.Debug().Msgf("error parsing url string: %s: %s\n", connectionString, err.Error())
		return nil, fmt.Errorf("error parsing url string: %s: %s\n", connectionString, err.Error())
	}

	redis.SetLogger(logger)
	log.Debug().Msgf("set up redis logger")
	return &Redis{
		connectionString: connectionString,
		logger:           logger,
		rOpts:            rOpts,
	}, nil
}

// WithConnectionString switches the mode of connection to redis to using
func WithConnectionString(s string) Options {
	return func(r *Redis) error {
		rOpts, err := redis.ParseURL(s)
		if err != nil {
			return fmt.Errorf("error parsing url string: %s: %s\n", s, err.Error())
		}
		r.conn = redis.NewClient(rOpts)
		return nil
	}
}

func (r *Redis) Connect(ctx context.Context) (bool, error) {
	log := r.logger.Logger()

	// apply options
	//if len(opts) > 0 {
	//	for i := 0; i < len(opts); i++ {
	//		err := opts[i](r)
	//		if err != nil {
	//			return false, err
	//		}
	//	}
	//	log.Debug().Msg("applied options successfully")
	//} else {
	//	log.Info().Msgf("no options provided moving ahead with default Redis connection config: %s", DefaultRedisConnectionString)
	//}

	// if connection isn't already created or I cannot ping the server successfully, then I'll create the connection using the default config.
	if r.conn == nil {
		r.conn = redis.NewClient(&redis.Options{
			Network:               "",
			Addr:                  "localhost:6379",
			ClientName:            "",
			Dialer:                nil,
			OnConnect:             nil,
			Protocol:              0,
			Username:              "",
			Password:              "",
			CredentialsProvider:   nil,
			DB:                    0,
			MaxRetries:            0,
			MinRetryBackoff:       0,
			MaxRetryBackoff:       0,
			DialTimeout:           0,
			ReadTimeout:           0,
			WriteTimeout:          0,
			ContextTimeoutEnabled: false,
			PoolFIFO:              false,
			PoolSize:              0,
			PoolTimeout:           0,
			MinIdleConns:          0,
			MaxIdleConns:          0,
			MaxActiveConns:        0,
			ConnMaxIdleTime:       0,
			ConnMaxLifetime:       0,
			TLSConfig:             nil,
			Limiter:               nil,
			DisableIndentity:      false,
			IdentitySuffix:        "",
		})
	}
	log.Debug().Msgf("created the client: %#v\n", *r.conn)

	b, err := r.Ping(ctx)
	if err != nil || !b {
		log.Debug().Msgf("encountered error pinging redis server %s: %v\n", r.connectionString, b)
		return false, err
	}
	log.Debug().Msgf("successfully pinged redis server: %s\n", r.connectionString)

	return true, nil
}

// Ping sends a ping message to the redis server to check the connection health
func (r *Redis) Ping(ctx context.Context) (bool, error) {
	log := r.logger.Logger()
	if err := r.Client().Ping(ctx).Err(); err != nil {
		log.Debug().Msgf("redis ping: unsuccessful")
		return false, err
	}
	log.Debug().Msgf("redis ping: successful")
	return true, nil
}

// Client returns the Redis client to the caller
func (r *Redis) Client() *redis.Client {
	return r.conn
}

// Store stores a key/value pair in the database.
func (r *Redis) Store(ctx context.Context, id string, token any) (err error) {
	log := r.logger.Logger()

	// ensure that token type is string
	var tokenStr string
	var b bool
	if b, tokenStr = InterfaceIsString(token); !b {
		log.Error().Msgf(ErrTokenTypeNotString)
		return fmt.Errorf(ErrTokenTypeNotString)
	}

	// check if key already exists
	if r.IsExists(ctx, tokenStr) {
		log.Error().Msgf("key already exists")
		return err
	}

	// check if key already exists
	status, err := r.Client().Set(ctx, id, tokenStr, DefaultTTL).Result()
	if err != nil {
		log.Error().Msgf(ErrWithOperation, err.Error())
		return err
	}
	if status != RedisStatusOkay {
		log.Debug().Msgf("did not receive an \"OK\" response from redis, received %s", RedisStatusOkay)
		return fmt.Errorf("did not receive an \"OK\" response from redis, received %s", RedisStatusOkay)
	}
	log.Debug().Msg(OperationSuccessful)
	return nil
}

// IsExists checks if a key already exists in redis
func (r *Redis) IsExists(ctx context.Context, key string) bool {
	_, err := r.Retrieve(ctx, key)
	if err == nil {
		return true
	}
	return false
}

// Retrieve retrieves a key/value pair from the database.
func (r *Redis) Retrieve(ctx context.Context, id string) (string, error) {
	val, err := r.Client().Get(ctx, id).Result()
	if err != nil {
		log.Error().Msgf(ErrWithOperation, err.Error())
		return val, err
	}
	log.Debug().Msg(OperationSuccessful)
	return val, nil
}

// RetrieveAll retrieves all the key/value pairs currently stored in the database.
func (r *Redis) RetrieveAll(ctx context.Context) (map[string]string, error) {
	log := r.logger.Logger()
	keys := r.conn.Keys(ctx, "*")
	l := len(keys.Val())
	// allocate map of size l
	kv := make(map[string]string, l)
	if l < 1 {
		log.Error().Msgf("length of retrieved keys is less than zero. it is: %d", l)
		return kv, fmt.Errorf("length of retrieved keys is less than zero. it is: %d", l)
	}

	for i := 0; i < l; i++ {
		kv[keys.Val()[i]] = r.conn.Get(ctx, keys.Val()[i]).Val()
	}
	log.Debug().Msgf("parsed database contents into map")

	return kv, nil
}

// Delete deletes a key/value pair identified by key
func (r *Redis) Delete(ctx context.Context, id string) (bool, error) {
	log := r.logger.Logger()
	n, err := r.conn.Del(ctx, id).Result()
	log.Info().Msgf("deleted %d number of keys", n)
	if err != nil {
		log.Error().Msgf("error occurred while deleting key %s: %s\n", id, err.Error())
		return true, fmt.Errorf("error occurred while deleting key %s: %s\n", id, err.Error())
	}
	return true, nil
}

// Patch replaces the value of a key in the redis DB
func (r *Redis) Patch(ctx context.Context, id string, token any) (bool, error) {
	log := r.logger.Logger()
	// check if key already exists
	if r.IsExists(ctx, id) {
		log.Error().Msgf("key already exists. patching")
	}

	var value string

	// ensure that token type is string
	var b bool
	if b, value = InterfaceIsString(token); !b {
		log.Error().Msgf(ErrTokenTypeNotString)
		return false, fmt.Errorf(ErrTokenTypeNotString)
	}

	status, err := r.Client().Set(ctx, id, value, DefaultTTL).Result()
	if err != nil {
		log.Error().Msgf(ErrWithOperation, err.Error())
		return false, err
	}
	if status != RedisStatusOkay {
		log.Debug().Msgf("did not receive an \"OK\" response from redis, received %s", RedisStatusOkay)
		return false, fmt.Errorf("did not receive an \"OK\" response from redis, received %s", RedisStatusOkay)
	}
	log.Debug().Msg(OperationSuccessful)
	return true, nil
}

// Flush clears the content of the current redis DB
func (r *Redis) Flush(ctx context.Context) (bool, error) {
	err := r.Client().FlushDB(ctx)
	if err.Err() != nil {
		log.Error().Msgf(ErrWithOperation, err.Err())
		return false, err.Err()
	}
	log.Debug().Msg(OperationSuccessful)
	return true, nil
}

// Close closes the redis connection
func (r *Redis) Close(ctx context.Context) error {
	return r.Client().Close()
}
