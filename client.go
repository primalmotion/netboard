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
	RunE: func(cmd *cobra.Command, args []string) error {

		addr := viper.GetString("url")
		certPath := viper.GetString("cert")
		certKeyPath := viper.GetString("cert-key")
		certKeyPass := viper.GetString("cert-key-pass")
		serverCAPath := viper.GetString("server-ca")
		skipVerify := viper.GetBool("insecure-skip-verify")
		runCmd := viper.GetString("cmd")
		runCmdArgs := viper.GetStringSlice("cmd-arg")

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
	clientCmd.Flags().StringP("url", "u", "https://127.0.0.1:8989", "The address of the netboard server")
	clientCmd.Flags().StringP("cert", "c", "", "Path to the client public key")
	clientCmd.Flags().StringP("cert-key", "k", "", "Path to the client private key")
	clientCmd.Flags().StringP("cert-key-pass", "p", "", "Optional client key passphrase")
	clientCmd.Flags().StringP("client-ca", "C", "", "Path to the server certificate CA")
	clientCmd.Flags().Bool("insecure-skip-verify", false, "Skip server CA validation. this is not secure")
	clientCmd.Flags().String("cmd", "wl-copy", "The command to run on new paste arrival")
	clientCmd.Flags().StringSlice("cmd-arg", nil, "Additional arguments to provide to cmd")
}
