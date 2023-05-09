package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"git.sr.st/~primalmotion/netboard/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.aporeto.io/tg/tglib"
	"golang.design/x/clipboard"
)

var listenCmd = &cobra.Command{
	Use:           "listen",
	Short:         "Sync data between clipboard and server",
	Args:          cobra.MaximumNArgs(0),
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		addr := viper.GetString("listen.url")
		certPath := os.ExpandEnv(viper.GetString("listen.cert"))
		certKeyPath := os.ExpandEnv(viper.GetString("listen.cert-key"))
		certKeyPass := viper.GetString("listen.cert-key-pass")
		serverCAPath := os.ExpandEnv(viper.GetString("listen.server-ca"))
		skipVerify := viper.GetBool("listen.insecure-skip-verify")

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

		watchChan := clipboard.Watch(cmd.Context(), clipboard.FmtText)
		listenChan := client.Listen(cmd.Context(), addr, tlsConf)

		for {
			select {
			case data := <-watchChan:
				if err := client.Copy(bytes.NewBuffer(data), addr, tlsConf); err != nil {
					log.Printf("error sending data: %s", err)
				}

			case data := <-listenChan:
				currentData := clipboard.Read(clipboard.FmtText)
				if !bytes.Equal(currentData, data) {
					log.Println("local clipboard updated")
					clipboard.Write(clipboard.FmtText, data)
				}

			case <-cmd.Context().Done():
				return nil
			}
		}
	},
}

func init() {
	listenCmd.Flags().StringP("listen.url", "u", "https://127.0.0.1:8989", "The address of the netboard server")
	viper.BindPFlag("listen.url", serverCmd.Flags().Lookup("url"))

	listenCmd.Flags().StringP("listen.cert", "c", "", "Path to the client public key")
	viper.BindPFlag("listen.cert", serverCmd.Flags().Lookup("cert"))

	listenCmd.Flags().StringP("listen.cert-key", "k", "", "Path to the client private key")
	viper.BindPFlag("listen.cert-key", serverCmd.Flags().Lookup("cert-key"))

	listenCmd.Flags().StringP("listen.cert-key-pass", "p", "", "Optional client key passphrase")
	viper.BindPFlag("listen.cert-key-pass", serverCmd.Flags().Lookup("cert-key-pass"))

	listenCmd.Flags().StringP("listen.server-ca", "C", "", "Path to the server certificate CA")
	viper.BindPFlag("listen.server-ca", serverCmd.Flags().Lookup("server-ca"))

	listenCmd.Flags().Bool("listen.insecure-skip-verify", false, "Skip server CA validation. this is not secure")
	viper.BindPFlag("listen.insecure-skip-verify", serverCmd.Flags().Lookup("insecure-skip-verify"))
}
