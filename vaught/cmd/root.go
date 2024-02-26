/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package cmd

import (
	"fmt"
	del "github.com/dark-enstein/vault/vaught/cmd/delete"
	"github.com/dark-enstein/vault/vaught/cmd/initer"
	"github.com/dark-enstein/vault/vaught/cmd/list"
	"github.com/dark-enstein/vault/vaught/cmd/peek"
	"github.com/dark-enstein/vault/vaught/cmd/peel"
	"github.com/dark-enstein/vault/vaught/cmd/service"
	"github.com/dark-enstein/vault/vaught/cmd/store"
	"os"

	"github.com/spf13/cobra"
)

const (
	FlagDebug = "debug"
)

type RootOptions struct {
	debug bool
}

// NewRootCmd represents the base command when called without any subcommands
func NewRootCmd() *cobra.Command {

	rop := &RootOptions{}

	rootCmd := &cobra.Command{
		Use:   "vault",
		Short: "Vault securely manages secrets and tokens",
		Long: `Vault is a command-line tool designed for secure management of secrets, tokens, and other sensitive data. Still experimental.
It provides a range of functionalities including storing, retrieving, deleting, and peeking at encrypted tokens, 
as well as initializing configurations for various storage backends.

With Vault, you can easily interact with different storage options such as local file systems, Redis servers, 
or in-memory storage, ensuring your sensitive data is managed securely and efficiently.

Examples:
  - Initialize Vault with a specific storage backend:
    vault init --store redis --connectionString "redis://user:password@localhost:6379"

  - Store a new secret token:
    vault store --id "myTokenID" [ --secret <sensitive value> | --secret-file <path to file containing secret> | --stdin <from stdin stream> ]

  - Only retrieve an encrypted token:
    vault peek --id "myTokenID"

  - Retrieve and decrypt a token:
    vault peel --id "myTokenID"

  - Delete a stored token:
    vault delete --id "myTokenID"

  - List all stored tokens:
    vault list

  To run vault as a service:
    vault service run [--port <port>]

Use "vault [command] --help" for more information about a command.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to Vault! Use 'vault [command] --help' for more information on a specific command.")
		},
	}

	rootCmd.AddCommand(srv.ServiceCmd)
	rootCmd.AddCommand(store.NewStoreCmd())
	rootCmd.AddCommand(peek.NewPeekCmd())
	rootCmd.AddCommand(peel.NewPeelCmd())
	rootCmd.AddCommand(list.NewListCmd())
	rootCmd.AddCommand(del.NewDeleteCmd())
	rootCmd.AddCommand(initer.NewInitCmd())
	rootCmd.PersistentFlags().BoolVarP(&rop.debug, FlagDebug, "d", false, "Enable or disable debug mode.")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
