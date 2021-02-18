package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var serviceAccountToken string

func readToken() {
	b, err := ioutil.ReadFile("/var/run/secrets/tokens/api-token")
	if err != nil {
		log.Fatalf("Cannot read token %s", err)
	}
	serviceAccountToken = string(b)
	log.Print("Refreshing SA token")
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	connectionString := os.Getenv("SERVICE_2_CONNECTION_STRING")
	if len(connectionString) == 0 {
		log.Fatal("SERVICE_2_CONNECTION_STRING must be set")
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", connectionString, nil)
	if err != nil {
		log.Fatal("Request cannot be created")
	}
	req.Header.Add("X-Client-Id", serviceAccountToken)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

func main() {

	readToken()

	ticker := time.NewTicker(5 * time.Minute)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				readToken()
			}
		}
	}()

	http.HandleFunc("/", handleIndex)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatal("PORT must be specified!")
	}
	http.ListenAndServe(port, nil)

	ticker.Stop()
	done <- true
}
