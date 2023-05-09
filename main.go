package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfgName string
)

var (
	version = "v0.0.0"
	commit  = "dev"
)

func main() {

	cobra.OnInitialize(initCobra)

	rootCmd := &cobra.Command{
		Use:              "netboard",
		Short:            "Simple and secure network clipboard sharing engine",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
	}

	rootCmd.AddCommand(
		serverCmd,
		clientCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func initCobra() {

	viper.SetEnvPrefix("netboard")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln("unable to find home dir: ", err)
	}

	if cfgFile == "" {
		cfgFile = os.Getenv("NETBOARD_CONFIG")
	}

	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			log.Fatalln("config file does not exist", err)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigFile(cfgFile)

		if err = viper.ReadInConfig(); err != nil {
			log.Fatalln("unable to read config", cfgFile)
		}

		return
	}

	viper.AddConfigPath(path.Join(home, ".config", "netboard"))
	viper.AddConfigPath("/usr/local/etc/netboard")
	viper.AddConfigPath("/etc/netboard")

	if cfgName == "" {
		cfgName = os.Getenv("NETBOARD_CONFIG_NAME")
	}

	if cfgName == "" {
		cfgName = "config"
	}

	viper.SetConfigName(cfgName)

	if err = viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			log.Fatalln("unable to read config:", err)
		}
	}
}
