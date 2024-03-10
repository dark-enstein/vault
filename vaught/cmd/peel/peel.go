/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package peel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dark-enstein/vault/internal/model"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/dark-enstein/vault/vaught/cmd/helper"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

type PeelOptions struct {
	id    string
	debug bool
}

// NewPeelCmd represents the cli command
func NewPeelCmd() *cobra.Command {

	pop := &PeelOptions{}

	peelCmd := &cobra.Command{
		Use:   "peel",
		Short: "Retrieves and decrypts a specific token by ID from the vault",
		Long: `The 'peel' command is designed to fetch and decrypt a specific token from the vault, using its unique identifier. 
This operation is critical for accessing encrypted data stored within the vault in a secure manner.

Usage:

  vault peel --id <token-id>

Substitute '<token-id>' with the actual ID of the token you need to access. Upon successful execution, this command will return the decrypted data associated with the token, ensuring secure access to sensitive information.

Supported Features:
- Secure retrieval: Ensures that the token is fetched securely and remains encrypted until it's safely within the application's context.
- On-demand decryption: Decrypts the token only when necessary, maintaining the confidentiality of the stored information.

Examples:
Decrypt and retrieve token data:
  vault peel --id 1234abcd

Make sure to run 'vault init' before attempting to peel a token, to ensure that the vault is properly configured and ready for secure operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Peeking record with ID %s\n", pop.id)
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Error().Msgf("error retrieving persistent flag: %s: %w", "debug", err)
			}

			ctx := context.Background()
			var logger = vlog.New(debug)

			jsonByte, err := pop.Run(ctx, logger)
			if err != nil {
				if errors.Is(err, helper.ErrConfigEmpty) || errors.Is(err, helper.ErrStoreTypeEmpty) {
					fmt.Println("config empty run `vault init` first. see more by running `vault init --help`")
					os.Exit(1)
				}
			}

			fmt.Println("All Tokens:")
			fmt.Println(string(jsonByte))
		},
	}

	peelCmd.Flags().StringVarP(&pop.id, "id", "i", "", "specify token ID to be peeled")
	return peelCmd
}

func (pop *PeelOptions) Run(ctx context.Context, logger *vlog.Logger) ([]byte, error) {
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

	token, err := manager.GetTokenByID(ctx, pop.id)
	if err != nil {
		logger.Logger().Fatal().Msgf("error retrieving token: %s", err)
		return nil, err
	}

	var children []*model.ChildReceipt

	b, decrypted, err := manager.Detokenize(ctx, pop.id, token.Data[0].Value)
	if err != nil {
		logger.Logger().Fatal().Msgf("error decrypting token: %s", err)
		return nil, err
	}

	children = append(children, &model.ChildReceipt{
		Key: pop.id,
		Value: &model.ChildResp{
			Found: b,
			Datum: decrypted,
		},
	})

	jsonByte, err := json.Marshal(children)
	if err != nil {
		logger.Logger().Fatal().Msgf("error marshalling decrypted token into json: %s", err)
		return nil, err
	}

	return jsonByte, nil
}
