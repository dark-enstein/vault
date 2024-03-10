/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dark-enstein/vault/internal/model"
	intstore "github.com/dark-enstein/vault/pkg/store"
	"github.com/dark-enstein/vault/pkg/vlog"
	"github.com/dark-enstein/vault/vaught/cmd/helper"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
)

type StoreOptions struct {
	id         string
	secretFlag string
	secret     string
	file       string
	stdin      bool
	stdinBuf   []byte
	debug      bool
	cmd        *cobra.Command
}

const (
	FlagID         = "id"
	FlagValue      = "secret"
	FlagSecretFile = "secret-file"
	FlagStdin      = "stdin"
	ErrBugs        = "BUG ERROR: %s. Please report this bug by filing an issue here %s. Thank you very much."
	IssueLink      = "" // TODO: fill it in
)

// NewStoreCmd represents the cli command
func NewStoreCmd() *cobra.Command {

	sop := &StoreOptions{}

	storeCmd := &cobra.Command{
		Use:   "store",
		Short: "Stores a new token with the specified value",
		Long: `The 'store' command securely stores a new token in the vault with a specified value, identified by a unique ID. 
This command is essential for adding new secrets to the vault, providing a safe way to manage sensitive information.

Usage:

  vault store --id <token-id> [ --secret <sensitive value> | --secret-file <path to file containing secret> | --stdin <from stdin stream> ] 

Replace '<token-id>' with the unique identifier for the new token, and '<secret-value>' with the actual secret information you wish to store. The command securely processes and stores the token in the configured storage backend, ensuring the confidentiality and integrity of your secret data.

Examples:
Store a new token:
  A. With secret value
  vault store --id "1234abcd" --secret "mySecretData"

  B. With secret from stdin
  echo $SECRET_STUFF | vault store --id "1234abcd" --stdin

  C. With secret from file
  vault store --id "1234abcd" --secret-file </path/to/secret/file>

Ensure to initialize the vault using 'vault init' before storing any tokens to set up the necessary configurations and storage backend.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Resolve persistent flags
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				log.Error().Msgf("error retrieving persistent flag: %s: %w", "debug", err)
			}
			logger := vlog.New(debug)
			ctx := context.Background()

			err = sop.Validate(ctx, logger)
			if err != nil {
				log.Fatal().Msgf("error validating flags: %s", err)
			}

			storeBytes, err := sop.Run(ctx, logger)
			if err != nil {
				log.Fatal().Msgf("%s", err)
			}

			sop.debug = debug
			if err != nil {
				// revisit this
				log.Fatal().Msgf(err.Error())
			}

			fmt.Printf("Stored token with id: %s\n", sop.id)
			fmt.Println(string(storeBytes))
		},
	}

	storeCmd.Flags().StringVarP(&sop.id, FlagID, "i", "", "specify token ID to be stored")
	storeCmd.Flags().StringVarP(&sop.secret, FlagValue, "s", "", "specify token ID to be stored")
	storeCmd.Flags().StringVarP(&sop.file, FlagSecretFile, "f", "", "specify token ID to be stored")
	storeCmd.Flags().BoolVarP(&sop.stdin, FlagStdin, "t", false, "specify token ID to be stored")
	storeCmd.MarkFlagsMutuallyExclusive(FlagStdin, FlagValue, FlagSecretFile)
	return storeCmd
}

func (sop *StoreOptions) Validate(ctx context.Context, logger *vlog.Logger) error {
	log := logger.Logger()

	// ensure that at least one of the data flags is set
	if !sop.stdin && len(sop.file) == 0 && len(sop.secret) == 0 {
		return errors.New("one of the flags [ --secret | --secret-file | --stdin ] must be set ")
	}

	// default flag is a naked secret. will change if file or stdin is set
	sop.secretFlag = FlagValue

	// if file path is set, read it into secret
	if len(sop.file) > 0 {
		if len(sop.secret) > 0 {
			return fmt.Errorf(ErrBugs, "trying to process secretfile, but secret already set", IssueLink)
		}

		sop.secretFlag = FlagSecretFile
		err := intstore.IsValidFile(sop.file, log)
		if err != nil {
			return err
		}

		fileBytes, err := os.ReadFile(sop.file)
		if err != nil {
			return err
		}

		if len(string(fileBytes)) == 0 {
			return errors.New("file passed in is empty, please pass in a file with the secret data")
		}

		// set secret as data read from file
		sop.secret = string(fileBytes)
	}

	// if stdin set, process stdin
	if sop.stdin {
		if len(sop.secret) > 0 {
			return fmt.Errorf(ErrBugs, "trying to process stdin, but secret already set", IssueLink)
		}
		sop.secretFlag = FlagStdin
		stdin := bytes.Buffer{}
		_, err := stdin.ReadFrom(os.Stdin)
		if err != nil {
			return fmt.Errorf(ErrBugs, err.Error(), IssueLink)
		}
		sop.stdinBuf = stdin.Bytes()

		// some validation?
		sop.secret = string(sop.stdinBuf)
	}

	// if secret is still empty, !bug
	if len(sop.secret) == 0 {
		return fmt.Errorf(ErrBugs, fmt.Sprintf("secret still empty, even after processing %s", sop.secretFlag), IssueLink)
	}

	return nil
}

func (sop *StoreOptions) Run(ctx context.Context, logger *vlog.Logger) ([]byte, error) {
	fmt.Println("Storing token with id:", sop.id)
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

	token, err := manager.Tokenize(ctx, sop.id, sop.secret)
	if err != nil {
		logger.Logger().Fatal().Msgf("error retrieving token: %s", err)
		return nil, err
	}

	// purge original secret from struct
	sop.secret = ""

	// capture token into struct
	tokenResp := &model.Child{
		Key:   sop.id,
		Value: token,
	}

	jsonByte, err := json.Marshal(tokenResp)
	if err != nil {
		logger.Logger().Fatal().Msgf("error marshalling token into json: %s", err)
		return nil, err
	}

	logger.Logger().Info().Msgf("Successfully stored id %s in the store.", sop.id)

	return jsonByte, nil
}
