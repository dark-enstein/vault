/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package initer

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/pkg/store"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/dark-enstein/vault/vaught/cmd/helper"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"time"
)

var (
	DefaultRootConfig = ""
)

const (
	FlagStoreLoc              = "fileLoc"
	FlagGobLoc                = "gobLoc"
	FlagStoreType             = "store"
	FlagRedisConnectionString = "connectionString"
)

type InitOptions struct {
	args            []string
	debug           bool
	storeStr        string
	gobLoc          string
	redisConnString string
	fileLoc         string
}

// NewInitCmd initializes the init command
func NewInitCmd() *cobra.Command {
	var opts = InitOptions{}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes your local system for on-demand secret encryption and storage using vault",
		Long: `The 'init' command sets up your local system for on-demand secret encryption and storage using vault.
This command supports multiple storage backends, allowing for flexible configurations based on the environment and requirements.

Supported storage options include:
- File: Utilizes the local file system for persistent storage.
- Gob: Employs GOB file storage for serialization of Go data structures.
- Redis: Connects to a Redis server for distributed storage and caching.
- In-memory map: Uses a concurrent-safe map for in-memory storage, ideal for temporary data and testing.

Examples:
Initialize the service with file storage:
  vault init --store file --fileLoc /path/to/store

Initialize the service with Redis:
  vault init --store redis --connectionString "redis://user:password@localhost:6379"

Each storage option offers specific flags for customization, providing the flexibility to adapt to various deployment scenarios.`,
		Run: func(cmd *cobra.Command, args []string) {
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Error().Msgf("error retrieving persistent flag: %s: %w", "debug", err)
			}
			logger := vlog.New(debug)
			ctx := context.Background()
			err = opts.Run(ctx, logger)

			// Resolve persistent flags

			opts.debug = debug
			if err != nil {
				// revisit this
				log.Fatal().Msgf(err.Error())
			}
		},
	}

	initCmd.Flags().StringVarP(&opts.storeStr, FlagStoreType, "s", "file", "Specify the storage backend for the service. Options: file, gob, redis, in-memory syncmap.")
	initCmd.Flags().StringVarP(&opts.redisConnString, FlagRedisConnectionString, "c", store.DefaultRedisConnectionString, "Specify the Redis connection string.")
	initCmd.Flags().StringVarP(&opts.gobLoc, FlagGobLoc, "g", helper.DefaultGobLoc, "Specify the disk location for the gob store.")
	initCmd.Flags().StringVarP(&opts.fileLoc, FlagStoreLoc, "f", helper.DefaultStoreLoc, "Specify the disk location for the file store.")
	initCmd.MarkFlagsMutuallyExclusive(FlagGobLoc, FlagStoreLoc, FlagRedisConnectionString)

	return initCmd
}

func (iop *InitOptions) Run(ctx context.Context, logger *vlog.Logger) error {
	fmt.Println("Initializing vault cli")
	var err error

	// check if config exists, override

	DefaultRootConfig, err = homedir.Expand(helper.DefaultRootConfig)
	if err != nil {
		logger.Logger().Error().Msgf("error occurred while setting up cli: %w", err)
		return err

	}

	ic := helper.InstanceConfig{
		ID:        xid.New().String(),
		CipherLoc: helper.DefaultCipherLoc,
		StoreType: iop.storeStr,
		Debug:     iop.debug,
		LastUse:   time.Now().UnixNano(),
	}

	// persist to disk at config loc
	err = ic.JsonEncode(helper.DefaultConfigLoc)
	if err != nil {
		logger.Logger().Error().Msgf("error occurred while setting up cli: %s", err)
		return err
	}

	logger.Logger().Info().Msgf("Successfully set up Vault CLI. You can begin using the other commands.")

	return nil
}
