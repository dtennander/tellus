// Package http serves the endpoint of the github web hook to Tellus.
package http

import (
	"encoding/json"
	"github.com/google/go-github/v24/github"
	"io"
	"log"
	"net/http"
	"tellus/tellus"
)

// StartHTTPServer creates a http server serving on the given port with the paths:
//   - "/api/github/webhook": The web hook for github.
//   - "/healthz": Status check for checking the health of the service.
func StartHTTPServer(port string, client *tellus.Client) {
	createRoutes(client)
	log.Printf("starting webserver on port: %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func createRoutes(tellusClient *tellus.Client) {
	http.HandleFunc("/api/github/webhook", withLogging(webhookHandler(tellusClient)))
	http.HandleFunc("/healthz", withLogging(healthzHandler()))
}

func healthzHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var rsp string
		_ = json.Unmarshal([]byte("ok"), rsp)
		bytes, e := json.Marshal("ok")
		if e != nil {
			http.Error(writer, e.Error(), 500)
		}
		_, _ = writer.Write(bytes)
	}
}

func webhookHandler(tellus *tellus.Client) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		eventType := r.Header.Get("X-GitHub-Event")
		switch eventType {
		case "pull_request":
			handlePullRequest(tellus, r.Body)
		case "push":
			handlePush(tellus, r.Body)
		case "check_run":
			handleCheckReRun(tellus, r.Body)
		}
	}
}

func handlePullRequest(tellusClient *tellus.Client, reader io.Reader) {
	var pullEvent github.PullRequestEvent
	err := json.NewDecoder(reader).Decode(&pullEvent)
	if err != nil {
		print(err.Error())
		return
	}
	err = tellusClient.NewPR(&pullEvent)
	if err != nil {
		print(err.Error())
		return
	}
	log.Printf("done handling pull_request event")
}

func handlePush(tellusClient *tellus.Client, reader io.Reader) {
	var pushEvent github.PushEvent
	err := json.NewDecoder(reader).Decode(&pushEvent)
	if err != nil {
		print(err.Error())
		return
	}
	err = tellusClient.NewPush(&pushEvent)
	if err != nil {
		print(err.Error())
		return
	}
	log.Printf("done handling push event")
}


func handleCheckReRun(client *tellus.Client, reader io.Reader) {
	var checkEvent github.CheckRunEvent
	err := json.NewDecoder(reader).Decode(&checkEvent)
	if err != nil {
		print(err.Error())
		return
	}
	err = client.CheckRunEvent(&checkEvent)
	if err != nil {
		print(err.Error())
		return
	}
	log.Printf("done handling run check event")
}
