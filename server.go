package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/primalmotion/netboard/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.aporeto.io/tg/tglib"
)

var serverCmd = &cobra.Command{
	Use:           "server",
	Short:         "Run the server",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		listenAddr := viper.GetString("server.listen")
		certPath := os.ExpandEnv(viper.GetString("server.cert"))
		certKeyPath := os.ExpandEnv(viper.GetString("server.cert-key"))
		certKeyPass := viper.GetString("server.cert-key-pass")
		clientCAPath := os.ExpandEnv(viper.GetString("server.client-ca"))

		log.Println("Server is listening on:", listenAddr)

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
	_ = viper.BindPFlag("server.listen", serverCmd.Flags().Lookup("listen"))

	serverCmd.Flags().StringP("cert", "c", "", "path to the server public key")
	_ = viper.BindPFlag("server.cert", serverCmd.Flags().Lookup("cert"))

	serverCmd.Flags().StringP("cert-key", "k", "", "path to the server private key")
	_ = viper.BindPFlag("server.cert-key", serverCmd.Flags().Lookup("cert-key"))

	serverCmd.Flags().StringP("cert-key-pass", "p", "", "optional server key passphrase")
	_ = viper.BindPFlag("server.cert-key-pass", serverCmd.Flags().Lookup("cert-key-pass"))

	serverCmd.Flags().StringP("client-ca", "C", "", "path to the client certificate CA")
	_ = viper.BindPFlag("server.client-ca", serverCmd.Flags().Lookup("client-ca"))
}
