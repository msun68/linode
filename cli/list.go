package cli

import (
	"context"
	"github.com/cheynewallace/tabby"
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
	instances, _ := linodeClient.ListInstances(context.Background(), nil)

	t := tabby.New()
	t.AddHeader("ID", "LABEL", "REGION", "TYPE", "IMAGE", "STATUS", "IP")

	for _, instance := range instances {
		var ips []string
		for _, ip := range instance.IPv4 {
			ips = append(ips, ip.String())
		}
		ips = append(ips, instance.IPv6)
		t.AddLine(instance.ID, instance.Label, instance.Region, instance.Type, instance.Image, instance.Status, ips)
	}

	t.Print()

	return nil
}
