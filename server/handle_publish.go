package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
)

func makePublishHandler(dispatch *dispatcher) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("unable to read body: %s", err),
				http.StatusBadRequest,
			)
			return
		}

		encoded := base64.RawURLEncoding.EncodeToString(data)
		encoded = encoded + ","

		id := computeID(r)
		log.Printf("dispatched data from: %s", id)

		dispatch.Dispatch(id, []byte(encoded))
		w.WriteHeader(http.StatusNoContent)
	}
}
