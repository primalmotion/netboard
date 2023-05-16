package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Publish sends the content of the given reader to the given url
// using the given tls config.
func Publish(data io.Reader, url string, tlsConfig *tls.Config) error {

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	r, err := http.NewRequest(http.MethodPost, url+"/publish", data)
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("unable to send request: %w", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("server rejected the request: %s", resp.Status)
	}

	log.Println("data dispatched")

	return nil
}
