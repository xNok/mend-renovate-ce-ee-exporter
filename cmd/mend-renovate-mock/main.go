package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/status", GetStatus).Methods("GET")

	http.ListenAndServe(":8010", router)
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 Internal Server Error"))
}

// Status was autogenerated from the testdata see pkg/metrics/testdata
type Status struct {
	BootDate time.Time `json:"bootDate"`
	Jobs     struct {
		LastEnqueueDate time.Time `json:"lastEnqueueDate"`
		LastJob         struct {
			InstallationID int `json:"installationId"`
			Params         struct {
			} `json:"params"`
			Priority   int    `json:"priority"`
			Reason     string `json:"reason"`
			Repository string `json:"repository"`
		} `json:"lastJob"`
		LastJobDispatchDate time.Time `json:"lastJobDispatchDate"`
		LastJobFinished     struct {
			Finished       time.Time `json:"finished"`
			InstallationID int       `json:"installationId"`
			Params         struct {
			} `json:"params"`
			Priority   int    `json:"priority"`
			Reason     string `json:"reason"`
			Repository string `json:"repository"`
		} `json:"lastJobFinished"`
		QueueLength        int `json:"queueLength"`
		TotalJobsProcessed int `json:"totalJobsProcessed"`
	} `json:"jobs"`
	JobsInProgress []struct {
		Repository string    `json:"repository"`
		Started    time.Time `json:"started"`
	} `json:"jobsInProgress"`
	Scheduler struct {
		Cron           string `json:"cron"`
		LastScheduling string `json:"lastScheduling"`
		Platform       string `json:"platform"`
	} `json:"scheduler"`
	Webhooks struct {
		LastWebhookReceived time.Time `json:"lastWebhookReceived"`
	} `json:"webhooks"`
	Worker struct {
		CurrentJob struct {
			InstallationID int `json:"installationId"`
			Params         struct {
			} `json:"params"`
			Priority   int    `json:"priority"`
			Reason     string `json:"reason"`
			Repository string `json:"repository"`
		} `json:"currentJob"`
		CurrentJobStart time.Time `json:"currentJobStart"`
		PreviousJob     struct {
			InstallationID int `json:"installationId"`
			Params         struct {
			} `json:"params"`
			Priority   int    `json:"priority"`
			Reason     string `json:"reason"`
			Repository string `json:"repository"`
		} `json:"previousJob"`
		PreviousJobStart       time.Time `json:"previousJobStart"`
		RemediateServerEnabled bool      `json:"remediateServerEnabled"`
	} `json:"worker"`
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
	// Recipe object that will be populated from json payload
	var status Status
	log.Info("GetStatus")

	err := faker.FakeData(&status)
	if err != nil {
		fmt.Println(err)
	}

	jsonBytes, err := json.Marshal(status)
	if err != nil {
		InternalServerErrorHandler(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
