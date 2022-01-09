package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/devopsfaith/bloomfilter/rpc/client"
	"github.com/gorilla/mux"
)

var key = "jti"

type Token struct {
	Jti string `json:"jti"`
}

func main() {
	server := "localhost:9999"

	c, err := client.New(server)
	if err != nil {
		log.Println("unable to create the rpc client:", err.Error())
		return
	}
	defer c.Close()

	router := mux.NewRouter()
	// Create
	router.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		addToken(w, r, c)
	}).Methods("POST")

	// Check
	router.HandleFunc("/check/{tokenId}", func(w http.ResponseWriter, r *http.Request) {
		checkToken(w, r, c)
	}).Methods("GET")

	log.Printf("Starting server...")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func addToken(w http.ResponseWriter, r *http.Request, bloomfilter *client.Bloomfilter) {
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(requestDump))

	var token Token
	json.NewDecoder(r.Body).Decode(&token)

	subject := key + "-" + token.Jti
	bloomfilter.Add([]byte(subject))
	log.Printf("adding [%s] %s", key, subject)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func checkToken(w http.ResponseWriter, r *http.Request, bloomfilter *client.Bloomfilter) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	tokenID := params["tokenId"]
	subject := key + "-" + tokenID

	res, err := bloomfilter.Check([]byte(subject))
	if err != nil {
		log.Println("Unable to check:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("checking [%s] %s => %v", key, subject, res)

	json.NewEncoder(w).Encode(res)
}
