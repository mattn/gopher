package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os/exec"

	"github.com/mattn/gopher"
)

type info struct {
	NumGopher int `json:"num_gopher"`
}

func findGopher() *gopher.Gopher {
	gophers := gopher.Lookup()
	if len(gophers) == 0 {
		return nil
	}
	return gophers[rand.Int()%len(gophers)]
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("assets")))
	http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&info{len(gopher.Lookup())})
	})
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			exec.Command("gopher").Start()
		}
	})
	http.HandleFunc("/say", func(w http.ResponseWriter, r *http.Request) {
		msg := r.FormValue("message")
		if r.Method != "POST" || msg == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if gopher := findGopher(); gopher != nil {
			gopher.Message(msg, "")
		}
	})
	http.ListenAndServe(":8080", nil)
}
