package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"git.sr.st/~primalmotion/netboard/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.aporeto.io/tg/tglib"
)

var clientCmd = &cobra.Command{
	Use:           "client",
	Short:         "Send data to server",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		addr := viper.GetString("client.url")
		certPath := os.ExpandEnv(viper.GetString("client.cert"))
		certKeyPath := os.ExpandEnv(viper.GetString("client.cert-key"))
		certKeyPass := viper.GetString("client.cert-key-pass")
		serverCAPath := os.ExpandEnv(viper.GetString("client.server-ca"))
		skipVerify := viper.GetBool("client.insecure-skip-verify")
		runCmd := viper.GetString("client.cmd")
		runCmdArgs := viper.GetStringSlice("client.cmd-arg")

		x509Cert, x509Key, err := tglib.ReadCertificatePEM(certPath, certKeyPath, certKeyPass)
		if err != nil {
			return fmt.Errorf("unable to read certificate: %w", err)
		}

		tlsCert, err := tglib.ToTLSCertificate(x509Cert, x509Key)
		if err != nil {
			return fmt.Errorf("unable to convert to tls certificate: %w", err)
		}

		var serverCAPool *x509.CertPool
		if serverCAPath != "" {
			serverCAData, err := os.ReadFile(serverCAPath)
			if err != nil {
				return fmt.Errorf("unable to read client certificate: %w", err)
			}
			serverCAPool = x509.NewCertPool()
			serverCAPool.AppendCertsFromPEM(serverCAData)
		} else {
			serverCAPool, err = x509.SystemCertPool()
			if err != nil {
				return fmt.Errorf("unable to prepare cert pool from system: %w", err)
			}
		}

		tlsConf := &tls.Config{
			Certificates:       []tls.Certificate{tlsCert},
			RootCAs:            serverCAPool,
			InsecureSkipVerify: skipVerify,
		}

		switch args[0] {
		case "copy":
			return client.Copy(addr, tlsConf)
		case "paste":

			for {
				if err := client.Paste(addr, tlsConf, runCmd, runCmdArgs...); err != nil {
					fmt.Fprintf(os.Stderr, "error during stream (retrying in 5sec): %s\n", err)
					time.Sleep(5 * time.Second)
				}
			}
		default:
			panic(fmt.Sprintf("unknown action :%s", args[0]))
		}
	},
}

func init() {
	clientCmd.Flags().StringP("client.url", "u", "https://127.0.0.1:8989", "The address of the netboard server")
	viper.BindPFlag("client.url", serverCmd.Flags().Lookup("url"))

	clientCmd.Flags().StringP("client.cert", "c", "", "Path to the client public key")
	viper.BindPFlag("client.cert", serverCmd.Flags().Lookup("cert"))

	clientCmd.Flags().StringP("client.cert-key", "k", "", "Path to the client private key")
	viper.BindPFlag("client.cert-key", serverCmd.Flags().Lookup("cert-key"))

	clientCmd.Flags().StringP("client.cert-key-pass", "p", "", "Optional client key passphrase")
	viper.BindPFlag("client.cert-key-pass", serverCmd.Flags().Lookup("cert-key-pass"))

	clientCmd.Flags().StringP("client.server-ca", "C", "", "Path to the server certificate CA")
	viper.BindPFlag("client.server-ca", serverCmd.Flags().Lookup("server-ca"))

	clientCmd.Flags().Bool("client.insecure-skip-verify", false, "Skip server CA validation. this is not secure")
	viper.BindPFlag("client.insecure-skip-verify", serverCmd.Flags().Lookup("insecure-skip-verify"))

	clientCmd.Flags().String("client.cmd", "wl-copy", "The command to run on new paste arrival")
	viper.BindPFlag("client.cmd", serverCmd.Flags().Lookup("cmd"))

	clientCmd.Flags().StringSlice("client.cmd-arg", nil, "Additional arguments to provide to cmd")
	viper.BindPFlag("client.cmd-arg", serverCmd.Flags().Lookup("cmd-arg"))
}
