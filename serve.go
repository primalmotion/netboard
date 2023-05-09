package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"git.sr.st/~primalmotion/netboard/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.aporeto.io/tg/tglib"
)

var serverCmd = &cobra.Command{
	Use:           "server",
	Short:         "Run the server",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		listenAddr := viper.GetString("listen")
		certPath := viper.GetString("cert")
		certKeyPath := viper.GetString("cert-key")
		certKeyPass := viper.GetString("cert-key-pass")
		clientCAPath := viper.GetString("client-ca")

		x509Cert, x509Key, err := tglib.ReadCertificatePEM(certPath, certKeyPath, certKeyPass)
		if err != nil {
			return fmt.Errorf("unable to read certificate: %w", err)
		}

		tlsCert, err := tglib.ToTLSCertificate(x509Cert, x509Key)
		if err != nil {
			return fmt.Errorf("unable to convert to tls certificate: %w", err)
		}

		clientCAData, err := os.ReadFile(clientCAPath)
		if err != nil {
			return fmt.Errorf("unable to read client certificate: %w", err)
		}

		clientCAPool := x509.NewCertPool()
		clientCAPool.AppendCertsFromPEM(clientCAData)

		tlsConf := &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCAPool,
		}

		return server.Serve(listenAddr, tlsConf)
	},
}

func init() {
	serverCmd.Flags().StringP("listen", "l", ":8989", "The listen address of the server")
	serverCmd.Flags().StringP("cert", "c", "", "path to the server public key")
	serverCmd.Flags().StringP("cert-key", "k", "", "path to the server private key")
	serverCmd.Flags().StringP("cert-key-pass", "p", "", "optional server key passphrase")
	serverCmd.Flags().StringP("client-ca", "C", "", "path to the client certificate CA")
}
