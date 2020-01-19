package cli

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/linode/linodego"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var (
	cfgFile      string
	linodeClient linodego.Client
)

// rootCmd represents the base command when called without any sub commands
var rootCmd = &cobra.Command{
	Use:               "linode",
	Short:             "Linode CLI",
	Long:              "This application is a tool to manage virtual machines hosted in Linode.",
	PersistentPreRunE: preRun,
	RunE:              run,
}

func init() {
	log.SetOutput(os.Stderr)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.linode.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".linode" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".linode")
	}

	viper.SetEnvPrefix("linode")
	viper.AutomaticEnv() // read in environment variables that match

	viper.SetDefault("debug", false)
	viper.SetDefault("personal_access_token", "")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func preRun(cmd *cobra.Command, args []string) error {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: viper.GetString("personal_access_token")})

	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linodeClient = linodego.NewClient(oauth2Client)
	linodeClient.SetDebug(viper.GetBool("debug"))

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	return cmd.Usage()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
