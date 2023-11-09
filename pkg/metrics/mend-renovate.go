package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/config"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/controller"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
)

const (
	mendRenovateStatusEndpoint     string           = "/api/status"
	TaskTypePullMendRenovateStatus schemas.TaskType = "TaskTypePullMendRenovateStatus"
)

// MendRenovateClient can be used to call Mend Renovate instance
type MendRenovateClient struct {
	URL   string
	Token string
}

// Status is the raw response of the API
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
		Cron           string      `json:"cron"`
		LastScheduling interface{} `json:"lastScheduling"`
		Platform       string      `json:"platform"`
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

// GetStatus call the status endpoint and collect the metrics
func (c *MendRenovateClient) GetStatus(ctx context.Context) (Status, error) {
	var status Status

	url, err := url.JoinPath(c.URL, mendRenovateStatusEndpoint)
	if err != nil {
		return status, fmt.Errorf("error building url: %v\n", err)
	}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", c.Token)
	if err != nil {
		return status, fmt.Errorf("error buulding http request: %v\n", err)
	}

	req = req.WithContext(ctx)
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return status, fmt.Errorf("error making http request: %v\n", err)
	}

	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		return status, fmt.Errorf("error decoding responce body: %v\n", err)
	}

	return status, nil
}

// MendRenovateController is used to handle task scheduling
type MendRenovateController struct {
	// Controller is the main controller handling scheduling
	Controller *controller.Controller
	// client
	client *MendRenovateClient
}

func NewMendRenovateController(c *controller.Controller) *MendRenovateController {
	return &MendRenovateController{
		Controller: c,
	}
}

// Configure set up the API metrics or sdk used to fetch the data.
func (c *MendRenovateController) Configure(ctx context.Context) {

	c.client = &MendRenovateClient{
		URL:   c.Controller.Config.Clients.MendRenovate.URL,
		Token: c.Controller.Config.Clients.MendRenovate.Token,
	}

	c.Controller.RegisterTasks(TaskTypePullMendRenovateStatus, c.taskHandlerPullStatus)
	c.Controller.Schedule(ctx, TaskTypePullMendRenovateStatus, config.SchedulerConfig(c.Controller.Config.Pull.Metrics))
	c.Controller.RegisterCollector(ctx, c.NewCollectors())
}

// taskHandlerPullStatus scrape men renovate metrics endpoint and store the relevant metrics
func (c *MendRenovateController) taskHandlerPullStatus(ctx context.Context) (err error) {
	defer c.Controller.UnqueueTask(ctx, TaskTypePullMendRenovateStatus, "_")
	defer c.Controller.TaskController.MonitorLastTaskScheduling(TaskTypePullMendRenovateStatus)

	status, err := c.client.GetStatus(ctx)
	if err != nil {
		return
	}

	controller.StoreSetMetric(
		ctx, c.Controller.Store, schemas.Metric{
			Kind:   MetricKindRenovateJobsQueueLength,
			Labels: nil,
			Value:  float64(status.Jobs.QueueLength),
		},
	)
	if err != nil {
		return err
	}

	return
}

// NewCollectors returns a new collector for resource exposed for this controller.
func (c *MendRenovateController) NewCollectors() controller.RegistryCollectors {
	return controller.RegistryCollectors{
		MetricKindRenovateJobsQueueLength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mre_renovate_jobs_queue_length",
				Help: "Number of Jobs in Renovate Queue",
			},
			[]string{},
		),
	}
}
