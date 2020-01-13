package cli

import (
	"context"

	"github.com/linode/linodego"
	"github.com/spf13/cobra"
)

var (
	createLabel          string
	createRegion         string
	createType           string
	createImage          string
	createLogin          string
	createAuthorizedKeys []string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new virtual machine",
	Long:  "This command creates a new virtual machine in Linode.",
	RunE:  create,
}

func init() {
	createCmd.Flags().StringVar(&createLabel, "label", "", "")
	createCmd.Flags().StringVar(&createRegion, "region", "us-west", "The region where the virtual machine will be located.")
	createCmd.Flags().StringVar(&createType, "type", "g6-nanode-1", "The type of the virtual machine you are creating.")
	createCmd.Flags().StringVar(&createImage, "image", "linode/ubuntu18.04", "An Image ID to deploy the Disk from.")
	createCmd.Flags().StringVar(&createLogin, "login", "", "")
	createCmd.Flags().StringArrayVar(&createAuthorizedKeys, "authorized-key", nil, "")
	createCmd.MarkFlagRequired("label")
	createCmd.MarkFlagRequired("login")
	createCmd.MarkFlagRequired("authorized-key")
	rootCmd.AddCommand(createCmd)
}

func create(cmd *cobra.Command, args []string) error {

	rootPass, err := GeneratePassword()

	if err != nil {
		return err
	}

	image, err := linodeClient.GetImage(context.Background(), createImage)

	if err != nil {
		return err
	}

	script, err := GenerateBootstrapScript(createLabel, createLogin, createAuthorizedKeys)

	if err != nil {
		return err
	}

	bootstrap, err := linodeClient.CreateStackscript(context.Background(), linodego.StackscriptCreateOptions{
		Label:    createLabel + "-bootstrap",
		Images:   []string{image.ID},
		IsPublic: false,
		Script:   script,
	})

	if err != nil {
		return err
	}

	falseBool := false

	instance, err := linodeClient.CreateInstance(context.Background(), linodego.InstanceCreateOptions{
		Region:    createRegion,
		Type:      createType,
		Label:     createLabel,
		PrivateIP: true,
		Booted:    &falseBool,
	})

	if err != nil {
		return err
	}

	disk, err := linodeClient.CreateInstanceDisk(context.Background(), instance.ID, linodego.InstanceDiskCreateOptions{
		Label:         image.Label + " Disk",
		Size:          instance.Specs.Disk,
		Image:         image.ID,
		RootPass:      rootPass,
		Filesystem:    "ext4",
		StackscriptID: bootstrap.ID,
	})

	_ = linodeClient.DeleteStackscript(context.Background(), bootstrap.ID)

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

	if err := linodeClient.BootInstance(context.Background(), instance.ID, config.ID); err != nil {
		_ = linodeClient.DeleteInstance(context.Background(), instance.ID)
		return err
	}

	for instance.Status != linodego.InstanceRunning {
		instance, err = linodeClient.GetInstance(context.Background(), instance.ID)
		if err != nil {
			_ = linodeClient.DeleteInstance(context.Background(), instance.ID)
			return err
		}
		PrintInstances(*instance)
	}

	return nil
}
