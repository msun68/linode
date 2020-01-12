package cli

import (
	"context"
	"github.com/linode/linodego"
	"github.com/spf13/cobra"
)

var (
	createRegion string
	createType   string
	createImage  string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <label>",
	Short: "Create a new virtual machine",
	Long:  "This command creates a new virtual machine in Linode.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  create,
}

func init() {
	createCmd.Flags().StringVar(&createRegion, "region", "us-west", "The region where the virtual machine will be located.")
	createCmd.Flags().StringVar(&createType, "type", "g6-nanode-1", "The type of the virtual machine you are creating.")
	createCmd.Flags().StringVar(&createImage, "image", "linode/ubuntu18.04", "An Image ID to deploy the Disk from.")
	rootCmd.AddCommand(createCmd)
}

func create(cmd *cobra.Command, args []string) error {

	rootPass, err := generatePassword()

	if err != nil {
		return err
	}

	image, err := linodeClient.GetImage(context.Background(), createImage)

	if err != nil {
		return err
	}

	falseBool := false

	instance, err := linodeClient.CreateInstance(context.Background(), linodego.InstanceCreateOptions{
		Region:    createRegion,
		Type:      createType,
		Label:     args[0],
		PrivateIP: true,
		Booted:    &falseBool,
	})

	if err != nil {
		return err
	}

	disk, err := linodeClient.CreateInstanceDisk(context.Background(), instance.ID, linodego.InstanceDiskCreateOptions{
		Label:          image.Label + " Disk",
		Size:           instance.Specs.Disk,
		Image:          image.ID,
		RootPass:       rootPass,
		Filesystem:     "ext4",
		AuthorizedKeys: nil,
	})

	if err != nil {
		_ = linodeClient.DeleteInstance(context.Background(), instance.ID)
		return err
	}

	config, err := linodeClient.CreateInstanceConfig(context.Background(), instance.ID, linodego.InstanceConfigCreateOptions{
		Label: "My " + image.Label + " Disk Profile",
		Devices: linodego.InstanceConfigDeviceMap{
			SDA: &linodego.InstanceConfigDevice{
				DiskID: disk.ID,
			},
		},
		Helpers: &linodego.InstanceConfigHelpers{
			UpdateDBDisabled:  true,
			Distro:            true,
			ModulesDep:        true,
			Network:           true,
			DevTmpFsAutomount: true,
		},
		Kernel: "linode/grub2",
	})

	if err != nil {
		_ = linodeClient.DeleteInstance(context.Background(), instance.ID)
		return err
	}

	err = linodeClient.BootInstance(context.Background(), instance.ID, config.ID)

	if err != nil {
		_ = linodeClient.DeleteInstance(context.Background(), instance.ID)
		return err
	}

	printInstances(*instance)

	return nil
}
