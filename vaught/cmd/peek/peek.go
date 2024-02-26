/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package peek

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dark-enstein/vault/internal/vlog"
	"github.com/dark-enstein/vault/vaught/cmd/helper"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type PeekOptions struct {
	id    string
	debug bool
}

// NewPeekCmd represents the CLI command for viewing a specific token
func NewPeekCmd() *cobra.Command {

	pop := &PeekOptions{}

	peekCmd := &cobra.Command{
		Use:   "peek",
		Short: "Displays details of a specific token by ID",
		Long: `The 'peek' command retrieves and decrypts a specific token stored in the vault, identified by its ID without decrypting it. 
This is useful for quickly viewing the encrypted version of a particular token without needing to list all tokens.

Usage:

  vault peek --id <token-id>

Replace '<token-id>' with the actual ID of the token you wish to view. The token's details will be displayed in JSON format, providing comprehensive information about the token's attributes.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Resolve persistent flags
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Error().Msgf("error retrieving persistent flag: %s: %w", "debug", err)
			}
			logger := vlog.New(debug)
			ctx := context.Background()
			bytes, err := pop.Run(ctx, logger)

			pop.debug = debug
			if err != nil {
				// revisit this
				log.Fatal().Msgf(err.Error())
			}

			fmt.Println("All Tokens:")
			fmt.Println(string(bytes))
		},
	}

	peekCmd.Flags().StringVarP(&pop.id, "id", "i", "", "specify token ID to be peeled")
	return peekCmd
}

func (pop *PeekOptions) Run(ctx context.Context, logger *vlog.Logger) ([]byte, error) {
	fmt.Println("Peeking at token with id:", pop.id)
	var err error

	// check if config exists, override
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

	token, err := manager.GetTokenByID(ctx, pop.id)
	if err != nil {
		logger.Logger().Fatal().Msgf("error retrieving token: %s", err)
		return nil, err
	}

	jsonByte, err := json.Marshal(token)
	if err != nil {
		logger.Logger().Fatal().Msgf("error marshalling token into json: %s", err)
		return nil, err
	}

	logger.Logger().Info().Msgf("Successfully set up Vault CLI. You can begin using the other commands.")

	return jsonByte, nil
}
