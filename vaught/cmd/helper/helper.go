package helper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dark-enstein/vault/internal/store"
	"github.com/dark-enstein/vault/internal/tokenize"
	"github.com/dark-enstein/vault/internal/vlog"
	"github.com/dark-enstein/vault/service"
	"os"
	"path/filepath"
)

var (
	ErrConfigEmpty      = errors.New("config loc is empty")
	ErrStoreTypeEmpty   = errors.New("store type is empty")
	ErrStoreTypeInvalid = errors.New("store type is invalid")
)

var (
	DefaultRootConfig = "~/.vault/cli"
	DefaultConfigLoc  = filepath.Join(DefaultRootConfig, "config")
	DefaultCipherLoc  = filepath.Join(DefaultRootConfig, ".cipher")
	DefaultStoreLoc   = filepath.Join(DefaultRootConfig, ".store")
	DefaultGobLoc     = filepath.Join(DefaultRootConfig, ".gob")
)

var logger = vlog.New(true)

type InstanceConfig struct {
	ID          string `json:"id"`
	CipherLoc   string `json:"cipher_loc"`
	StoreLoc    string `json:"store_loc"`
	RedisString string `json:"redis_string"`
	StoreType   string `json:"store_type"`
	Debug       bool   `json:"debug"`
	LastUse     int64  `json:"last_login"`
}

func NewInstanceConfig() *InstanceConfig {
	return &InstanceConfig{}
}

func (ic *InstanceConfig) Manager(ctx context.Context) (*tokenize.Manager, error) {
	if len(ic.StoreLoc) == 0 {
		return nil, ErrStoreTypeEmpty
	}
	log := logger.Logger()
	switch ic.StoreLoc {
	case service.STORE_FILE:
		log.Info().Msg("Using File storage")
		return tokenize.NewManager(ctx, logger, tokenize.WithStore(store.NewFile(ic.StoreLoc, logger))), nil
	case service.STORE_GOB:
		log.Info().Msg("Using Gob storage")
		gob, err := store.NewGob(ctx, ic.StoreLoc, logger, false)
		if err != nil {
			log.Fatal().Msgf("error while creating storage backend: %s", err)
		}
		return tokenize.NewManager(ctx, logger, tokenize.WithStore(gob)), nil
	case service.STORE_REDIS:
		log.Info().Msg("Using Redis storage")
		r, err := store.NewRedis(ic.RedisString, logger)
		if err != nil {
			log.Fatal().Msgf("error while creating storage backend: %s", err)
		}
		return tokenize.NewManager(ctx, logger, tokenize.WithStore(r)), nil
	case service.STORE_MAP:
		log.Info().Msg("Using In-memory map storage")
		return tokenize.NewManager(ctx, logger, tokenize.WithStore(store.NewSyncMap(ctx, logger))), nil
	default:
		return nil, ErrStoreTypeInvalid
	}
}

func (ic *InstanceConfig) JsonEncode(path string) error {
	log := logger.Logger()

	log.Info().Msgf("setting up vault config at location: %s", path)

	_, err := os.Stat(DefaultRootConfig)
	if os.IsNotExist(err) {
		err = os.MkdirAll(DefaultRootConfig, 0655)
		if err != nil {
			log.Error().Msgf("encountered error while setting up config dir %s: %s", path, err)
			return err
		}
	}

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Error().Msgf("encountered error while opening file at location %s: %s", path, err)
		return err
	}
	defer fd.Close()

	jsonBytes, err := json.Marshal(&ic)
	if err != nil {
		log.Error().Msgf("encountered error while encoding json %s: %s", path, err)
		return err
	}

	fmt.Println(string(jsonBytes))

	_, err = fd.Write(jsonBytes)
	if err != nil {
		log.Error().Msgf("encountered error while writing config to file location %s: %s", path, err)
		return err
	}

	return nil
}
func (ic *InstanceConfig) jsonDecode(path string) error {
	log := logger.Logger()

	log.Info().Msgf("setting up vault config at location: %s", path)

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		log.Error().Msgf("encountered error while reading config file at location %s: %s", path, err)
		return err
	}
	if len(fileBytes) == 0 {
		return ErrConfigEmpty
	}

	// reset instance config
	ic = &InstanceConfig{}

	err = json.Unmarshal(fileBytes, &ic)
	if err != nil {
		log.Error().Msgf("json config invalid: %s", err)
		return err
	}

	return nil
}

func (ic *InstanceConfig) JsonDecode() error {
	return ic.jsonDecode(DefaultConfigLoc)
}
