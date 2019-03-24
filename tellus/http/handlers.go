package http

import (
	"encoding/json"
	"github.com/google/go-github/v24/github"
	"io"
	"log"
	"net/http"
	"tellus/tellus"
)

func ServeHttpClient(port string, client *tellus.Client) {
	createRoutes(client)
	log.Printf("starting webserver on port: %s", port)
	err := http.ListenAndServe(":" + port, nil)
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
		eventType := r.Header.Get("X-GitHub-Event")
		switch eventType {
		case "pull_request":
			go func() {
				defer r.Body.Close()
				handlePullRequest(tellus, r.Body)
			}()
		case "push":
			go func() {
				defer r.Body.Close()
				handlePush(tellus, r.Body)
			}()
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

