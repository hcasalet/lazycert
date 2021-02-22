package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	// The beginning of LazyCert

	messageHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Here we start our POC of LazyCert!\n")
	}

	http.HandleFunc("/LazyCert", messageHandler)
        log.Println("Listing for requests at http://localhost:8000/LazyCert")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
