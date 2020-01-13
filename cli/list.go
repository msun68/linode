package cli

import (
	"context"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List virtual machines",
	Long:  "This command lists all of the virtual machines hosted in Linode.",
	RunE:  list,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) error {

	instances, err := linodeClient.ListInstances(context.Background(), nil)

	if err != nil {
		return err
	}

	PrintInstances(instances...)

	return nil
}
