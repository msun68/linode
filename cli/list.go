package cli

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/linode/linodego"
	"github.com/spf13/cobra"
)

var (
	listRegion string
	listTag    string
	listFormat string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List virtual machines",
	Long:  "This command lists all of the virtual machines hosted in Linode.",
	RunE:  list,
}

type filter struct {
	Region string `json:"region,omitempty"`
	Tag    string `json:"tags,omitempty"`
}

func init() {
	listCmd.Flags().StringVar(&listRegion, "region", "", "")
	listCmd.Flags().StringVar(&listTag, "tag", "", "")
	listCmd.Flags().StringVar(&listFormat, "format", "", "")
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) error {

	filterJson, _ := json.Marshal(filter{
		Region: listRegion,
		Tag:    listTag,
	})

	instances, err := linodeClient.ListInstances(context.Background(), &linodego.ListOptions{
		Filter: string(filterJson),
	})

	if err != nil {
		return err
	}

	var printOptions *PrintInstancesOptions

	fields := strings.Split(regexp.MustCompile(`\s+`).ReplaceAllString(listFormat, ""), ":")

	if fields[0] == "ansible" {
		printOptions = &PrintInstancesOptions{
			Type: "ansible",
		}
		if len(fields) > 1 {
			printOptions.Options = make(map[string]*string)
			for _, opts := range strings.Split(fields[1], ",") {
				printOptions.Options[opts] = nil
			}
		}
	}

	PrintInstances(instances, GetIPAddresses(), printOptions)

	return nil
}
