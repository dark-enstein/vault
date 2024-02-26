/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package del

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/internal/vlog"
	"github.com/dark-enstein/vault/vaught/cmd/helper"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

type DeleteOptions struct {
	id string
}

// NewDeleteCmd represents the cli command
func NewDeleteCmd() *cobra.Command {

	do := DeleteOptions{}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Deletes a specific token from the vault",
		Long: `The 'delete' command removes a specified token from the vault's storage, identified by its unique ID. This action is irreversible, emphasizing the need for caution.

Supported deletion operation:

- By ID: Targets a specific token for deletion based on its unique identifier.

Examples:
Delete a token by its ID:
  vault delete --id <token-id>

Ensure the correct token ID is specified to prevent unintended data loss. This command is designed for precise operation, allowing for the secure management and cleanup of stored tokens.

Each deletion operation requires the token ID to be specified, ensuring targeted and secure removal of sensitive information.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Deleting record with id:", do.id)
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Error().Msgf("error retrieving persistent flag: %s: %w", "debug", err)
			}

			ctx := context.Background()
			var logger = vlog.New(debug)

			err = do.Run(ctx, logger)
			if err != nil {
				if errors.Is(err, helper.ErrConfigEmpty) || errors.Is(err, helper.ErrStoreTypeEmpty) {
					fmt.Println("config empty run `vault init` first. see more by running `vault init --help`")
					os.Exit(1)
				}
			}
			fmt.Println("Deleted successfully")
		},
	}

	deleteCmd.Flags().StringVarP(&do.id, "id", "i", "", "specify token ID to be deleted.")
	return deleteCmd
}

func (do *DeleteOptions) Run(ctx context.Context, logger *vlog.Logger) error {
	fmt.Println("Initializing vault cli")
	var err error

	ic := helper.NewInstanceConfig()
	err = ic.JsonDecode()
	if err != nil {
		return err
	}

	// initialize token manager
	manager, err := ic.Manager(ctx)
	if err != nil {
		return err
	}

	_, err = manager.DeleteTokenByID(ctx, do.id)
	// ignore error, since delete op
	return nil
}
