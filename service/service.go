package service

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/internal/tokenize"
	"github.com/dark-enstein/vault/pkg/store"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

const (
	port        = "8080"
	STORE_REDIS = "redis"
	STORE_GOB   = "gob"
	STORE_FILE  = "file"
	STORE_MAP   = "map"
)

var (
	ErrInvalidRequestParameter = errors.New("invalid request parameter")
)

type Service struct {
	sc         *StartConfig
	manager    *tokenize.Manager
	srv        *http.Server
	mux        *http.ServeMux
	log        *vlog.Logger
	storeStr   string
	fileConfig struct {
		loc string
	}
	gobConfig struct {
		loc   string
		trunc bool
	}
	redisConfig struct {
		connectionString string
	}
	syncMapConfig struct{}
}

func New(ctx context.Context, log *vlog.Logger, opts ...Options) (*Service, error) {
	srv := &Service{sc: &StartConfig{port: port}, mux: http.NewServeMux(), log: log}

	// fill in the gaps in the struct
	for i := 0; i < len(opts); i++ {
		opts[i](srv)
	}

	if len(opts) == 0 {
		srv.SetDefaults()
	}

	// resolve store type
	store, err := srv.isInvalidStore(ctx)
	if err != nil {
		return nil, err
	}

	srv.manager = tokenize.NewManager(ctx, srv.log, tokenize.WithStore(store))

	log.Logger().Debug().Msg("generating service config")
	readTimeout := 10 * time.Second
	writeTimeout := 10 * time.Second
	srv.srv = &http.Server{
		Addr:                         ":" + port,
		Handler:                      nil,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  readTimeout,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 writeTimeout,
		IdleTimeout:                  0,
		MaxHeaderBytes:               1 << 20,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	log.Logger().Debug().Msgf("initialized service with settings:\n\taddress: %v\n\tread timeout: %v\n\twrite timeout: %v\n", ":"+port, readTimeout, writeTimeout)
	return srv, nil
}

type StartConfig struct {
	port string
}

func (s *Service) Port() string {
	return s.sc.port
}

func (s *Service) LoadHandlers(ctx context.Context) {
	for k, v := range *NewVaultHandler(ctx, s) {
		s.mux.HandleFunc(k, v)
	}
}

func (s *Service) Run(ctx context.Context) error {
	// load handlers into mux
	s.LoadHandlers(ctx)
	// set mux into server
	s.srv.Handler = s.mux
	// start server
	return s.srv.ListenAndServe()
}

// isInvalidStore resolves the underlying store struct based on the passed in info
func (s *Service) isInvalidStore(ctx context.Context) (store.Store, error) {
	// no-op
	if len(s.storeStr) == 0 {
		return nil, nil
	}

	var val store.Store
	var err error
	switch s.storeStr {
	case STORE_REDIS:
		val, err = store.NewRedis(s.redisConfig.connectionString, s.log)
	case STORE_FILE:
		val = store.NewFile(s.fileConfig.loc, s.log)
	case STORE_GOB:
		val, err = store.NewGob(ctx, s.gobConfig.loc, s.log, s.gobConfig.trunc)
	case STORE_MAP:
		val = store.NewSyncMap(ctx, s.log)
	default:
		return nil, fmt.Errorf("%w: input store invalid", ErrInvalidRequestParameter)
	}
	return val, err
}

func (s *Service) SetDefaults() {
	if len(s.storeStr) == 0 {
		s.storeStr = STORE_GOB
		s.gobConfig.loc = ".gob"
		s.gobConfig.trunc = true
	}
}

type Options func(*Service)

func WithFileLoc(loc string) Options {
	return func(s *Service) {
		s.fileConfig.loc = loc
	}
}

func WithGobLoc(loc string) Options {
	return func(s *Service) {
		s.gobConfig.loc = loc
	}
}

func WithGobTrunc() Options {
	return func(s *Service) {
		s.gobConfig.trunc = true
	}
}

func WithRedisConnString(connStr string) Options {
	return func(s *Service) {
		s.redisConfig.connectionString = connStr
	}
}

func WithStoreStr(storeStr string) Options {
	return func(s *Service) {
		s.storeStr = storeStr
	}
}
