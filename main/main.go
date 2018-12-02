package main

import (
	"fmt"
	"github.com/supreethrao/support-bot/rota"
	"log"
	"net/http"
)

func serve() {
	handleMembersList()
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func handleMembersList() {
	http.HandleFunc("/members", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, err := fmt.Fprint(w, rota.List())
			if err != nil {
				log.Fatal(err)
			}
		}
	})
}

func main() {
	serve()
}
