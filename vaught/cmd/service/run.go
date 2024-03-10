/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package srv

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/pkg/store"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/dark-enstein/vault/service"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// runCmd represents the service command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the service with the specified storage backend.",
	Long: `Starts the server and initializes the specified storage backend for data persistence. 
This command supports multiple storage backends, allowing flexible configurations based on the environment and requirements.

Supported storage options include:
- File: A local file system-based storage for persistence.
- Gob: A GOB file storage, offering serialization for Go data structures.
- Redis: A Redis server connection for distributed storage and caching.
- In-memory concurrent-safe map: A concurrent map for in-memory storage, suitable for temporary data and testing purposes.

Examples:
Run the service with file storage:
  vault service run --store file --fileLoc /path/to/store

Run the service with Redis:
  vault service run --store redis --connectionString "redis://user:password@localhost:6379"

Each storage option has its specific flags for customization, providing flexibility to adapt to various deployment scenarios.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing vault service")

		logger := vlog.New(true)

		ctx := context.Background()

		var srv *service.Service
		var err error
		switch storeStr {
		case service.STORE_FILE:
			log.Info().Msg("Using File storage")
			srv, err = service.New(ctx, logger, service.WithStoreStr(storeStr), service.WithFileLoc(fileLoc))
		case service.STORE_GOB:
			log.Info().Msg("Using Gob storage")
			srv, err = service.New(ctx, logger, service.WithStoreStr(storeStr), service.WithGobLoc(gobLoc))
		case service.STORE_REDIS:
			log.Info().Msg("Using Redis storage")
			srv, err = service.New(ctx, logger, service.WithStoreStr(storeStr), service.WithRedisConnString(redisConnString))
		case service.STORE_MAP:
			log.Info().Msg("Using In-memory map storage")
			srv, err = service.New(ctx, logger, service.WithStoreStr(storeStr))
		}
		if err != nil {
			logger.Logger().Fatal().Msgf("error while setting up service: %w", err)
		}

		if err := srv.Run(ctx); err != nil {
			logger.Logger().Fatal().Msgf("error while service is starting: %s\n", err.Error())
		}
	},
}

var port string
var storeStr string
var redisConnString string
var gobLoc string
var fileLoc string
var debug bool

func init() {

	// init flags
	runCmd.Flags().StringVarP(&port, "port", "p", "8080", "Specify port for service to listen on")
	runCmd.Flags().StringVarP(&storeStr, "store", "s", "file", "Specify which of the store you would like the service to connect to. Options: file, gob, redis, in-memory syncmap.")
	runCmd.Flags().StringVarP(&redisConnString, "connectionString", "c", store.DefaultRedisConnectionString, "Specify the connection string to redis")
	runCmd.Flags().StringVarP(&gobLoc, "gobLoc", "g", ".gob", "Specify the disk location of the gob store")
	runCmd.Flags().StringVarP(&fileLoc, "fileLoc", "f", ".store", "Specify the disk location of the file store")
	runCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Toggle debug mode")
}
