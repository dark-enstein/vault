/*
Copyright © 2024 Ayobami Bamigboye <ayo@greystein.com>
*/
package srv

import (
	"github.com/spf13/cobra"
)

// ServiceCmd represents the service command
var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manages vaught's application's services",
	Long: `Manages the various features of the vault service, providing commands to start in foreground/background, stop, and restarting. Coming soon: hot modification of configurations later. 
This command acts as a gateway for interacting with different aspects of vault's service functionality.

Examples of usage include:

- Starting a service in foreground or background

# coming soon
- Stopping a service (requires the service is started in background)
- Restarting a service [w]ith new parameters] (requires the service is started in background)
- Viewing the status of running services (requires the service is started in background)`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	ServiceCmd.AddCommand(runCmd)
}
