package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/primalmotion/netboard/cboard"
	"github.com/primalmotion/netboard/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.aporeto.io/tg/tglib"
)

var listenCmd = &cobra.Command{
	Use:           "listen",
	Short:         "Sync data between clipboard and server",
	Args:          cobra.MaximumNArgs(0),
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return err
		}
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		addr := viper.GetString("listen.url")
		certPath := os.ExpandEnv(viper.GetString("listen.cert"))
		certKeyPath := os.ExpandEnv(viper.GetString("listen.cert-key"))
		certKeyPass := viper.GetString("listen.cert-key-pass")
		serverCAPath := os.ExpandEnv(viper.GetString("listen.server-ca"))
		skipVerify := viper.GetBool("listen.insecure-skip-verify")
		mode := viper.GetString("listen.mode")

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

		var cb cboard.ClipboardManager

		switch mode {
		case "lib":
			cb, err = cboard.NewLibClipboardManager()
			if err != nil {
				log.Fatalf("unable to use lib mode: %s", err)
			}
			log.Println("using lib mode")
		case "wl-clipboard":
			cb, err = cboard.NewToolsClipboardManager()
			if err != nil {
				log.Fatalf("unable to use wl-clipboard mode: %s", err)
			}
			log.Println("using wl-clipboard mode")
		default:
			log.Fatalf("unknown mode %s", mode)
		}

		watchChan := cb.Watch(cmd.Context())
		listenChan := client.Listen(cmd.Context(), addr, tlsConf)

		for {
			select {
			case data := <-watchChan:
				log.Println("local clipboard changed. pushing")
				if err := client.Send(bytes.NewBuffer(data), addr, tlsConf); err != nil {
					log.Printf("error sending data: %s", err)
				}

			case data := <-listenChan:
				currentData, err := cb.Read()
				if err != nil {
					log.Printf("unable to read from local clipboard: %s", err)
					continue
				}

				if !bytes.Equal(currentData, data) {
					log.Println("local clipboard updated")
					if err := cb.Write(data); err != nil {
						log.Printf("unable to write to local clipboard: %s", err)
						continue
					}
				}

			case <-cmd.Context().Done():
				return nil
			}
		}
	},
}

func init() {
	listenCmd.Flags().StringP("url", "u", "https://127.0.0.1:8989", "The address of the netboard server")
	_ = viper.BindPFlag("listen.url", listenCmd.Flags().Lookup("url"))

	listenCmd.Flags().StringP("cert", "c", "", "Path to the client public key")
	_ = viper.BindPFlag("listen.cert", listenCmd.Flags().Lookup("cert"))

	listenCmd.Flags().StringP("cert-key", "k", "", "Path to the client private key")
	_ = viper.BindPFlag("listen.cert-key", listenCmd.Flags().Lookup("cert-key"))

	listenCmd.Flags().StringP("cert-key-pass", "p", "", "Optional client key passphrase")
	_ = viper.BindPFlag("listen.cert-key-pass", listenCmd.Flags().Lookup("cert-key-pass"))

	listenCmd.Flags().StringP("server-ca", "C", "", "Path to the server certificate CA")
	_ = viper.BindPFlag("listen.server-ca", listenCmd.Flags().Lookup("server-ca"))

	listenCmd.Flags().Bool("insecure-skip-verify", false, "Skip server CA validation. this is not secure")
	_ = viper.BindPFlag("listen.insecure-skip-verify", listenCmd.Flags().Lookup("insecure-skip-verify"))

	listenCmd.Flags().String("mode", "wl-clipboard", "Select the mode to handle clipboard. wl-clipboard or lib")
	_ = viper.BindPFlag("listen.mode", listenCmd.Flags().Lookup("mode"))
}
