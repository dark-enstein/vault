/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package list

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/dark-enstein/vault/vaught/cmd/helper"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

type ListOptions struct {
}

// NewListCmd represents the CLI command for listing all stored tokens
func NewListCmd() *cobra.Command {

	lop := ListOptions{}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all stored tokens in the vault",
		Long: `The 'list' command retrieves and displays all tokens currently stored in the vault. 
This command is useful for getting an overview of all the secrets managed by the vault system. 

Usage:

  vault list

This will output all the tokens stored, formatted as JSON for easy reading and integration with other tools. Ensure you have the appropriate permissions and the vault is correctly configured before running this command.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing records in vault")
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Error().Msgf("error retrieving persistent flag: %s: %w", "debug", err)
			}

			ctx := context.Background()
			logger := vlog.New(debug)

			bytes, err := lop.Run(ctx, logger)
			if err != nil {
				if errors.Is(err, helper.ErrConfigEmpty) || errors.Is(err, helper.ErrStoreTypeEmpty) {
					fmt.Println("config empty run `vault init` first. see more by running `vault init --help`")
					os.Exit(1)
				}
			}

			fmt.Println("All Tokens:")
			fmt.Println(string(bytes))
		},
	}

	return listCmd

}

func (lop *ListOptions) Run(ctx context.Context, logger *vlog.Logger) ([]byte, error) {
	fmt.Println("Initializing vault cli")
	var err error

	ic := helper.NewInstanceConfig()
	err = ic.JsonDecode()
	if err != nil {
		return nil, err
	}

	// initialize token manager
	manager, err := ic.Manager(ctx)
	if err != nil {
		logger.Logger().Debug().Msgf("error retrieving tokens from store: %s", err)
		return nil, err
	}

	tokens, err := manager.GetAllTokens(ctx)
	if err != nil {
		logger.Logger().Error().Msgf("error retrieving tokens from store: %s", err)
		return nil, err
	}

	bytesResult, err := json.Marshal(&tokens)
	if err != nil {
		logger.Logger().Error().Msgf("error marshalling tokens into json: %s", err)
		return nil, err
	}

	return bytesResult, nil
}
