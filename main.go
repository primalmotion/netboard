package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
				return err
			}
			return viper.BindPFlags(cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			if viper.GetBool("version") {
				fmt.Printf("netboard %s (%s)\n", version, commit)
				os.Exit(0)
			}
		},
	}
	rootCmd.Flags().Bool("version", false, "Show version")

	rootCmd.AddCommand(
		serverCmd,
		listenCmd,
	)

	mainCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	signalCh := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		cancelFunc()
		signal.Stop(signalCh)
		close(signalCh)
	}()

	if err := rootCmd.ExecuteContext(mainCtx); err != nil {
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
