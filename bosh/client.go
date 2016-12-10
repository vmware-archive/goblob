package bosh

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"bytes"
)

var (
	transport  = http.DefaultTransport
	bodyReader = ioutil.ReadAll
)

type Config struct {
	URL                 string
	Username            string
	Password            string
	TaskPollingInterval time.Duration
	AllowInsecureSSL    bool
}

type Client struct {
	config Config
}

type Task struct {
	Id     int
	State  string
	Result string
}

type TaskOutput struct {
	Time     int64
	Error    TaskError
	Stage    string
	Tags     []string
	Total    int
	Task     string
	Index    int
	State    string
	Progress int
}

type TaskError struct {
	Code    int
	Message string
}

type VM struct {
	Index   int      `json:"index"`
	State   string   `json:"job_state"`
	JobName string   `json:"job_name"`
	IPs     []string `json:"ips"`
}

type Option struct {

}

func SetBodyReader(r func(io.Reader) ([]byte, error)) {
	bodyReader = r
}

func ResetBodyReader() {
	bodyReader = ioutil.ReadAll
}

func NewClient(config Config) Client {
	if config.TaskPollingInterval == time.Duration(0) {
		config.TaskPollingInterval = 5 * time.Second
	}

	if config.AllowInsecureSSL {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return Client{
		config: config,
	}
}

func (c Client) GetVMs(name string) ([]VM, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments/%s/vms?format=full", c.config.URL, name), nil)
	if err != nil {
		return []VM{}, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	response, err := transport.RoundTrip(request)
	if err != nil {
		return []VM{}, err
	}

	if response.StatusCode != http.StatusFound {
		body, err := bodyReader(response.Body)
		if err != nil {
			return []VM{}, err
		}
		defer response.Body.Close()

		return []VM{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	location := response.Header.Get("Location")

	_, err = c.checkTaskStatus(location)
	if err != nil {
		return []VM{}, err
	}

	location, err = c.rewriteURL(location)
	if err != nil {
		return []VM{}, err
	}

	request, err = http.NewRequest("GET", fmt.Sprintf("%s/output?type=result", location), nil)
	if err != nil {
		return []VM{}, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	response, err = transport.RoundTrip(request)
	if err != nil {
		return []VM{}, err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return []VM{}, err
	}
	defer response.Body.Close()

	body = bytes.TrimSpace(body)
	parts := bytes.Split(body, []byte("\n"))

	var vms []VM
	for _, part := range parts {
		var vm VM
		err = json.Unmarshal(part, &vm)
		if err != nil {
			return vms, err
		}

		vms = append(vms, vm)
	}

	return vms, nil
}

func (c Client) Start(deployment, job string, index int) error {
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/deployments/%s/jobs/%s/%d?state=start", c.config.URL, deployment, job, index), bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	request.Header.Set("Content-Type", "text/yaml")
	response, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusFound {
		responseBody, err := bodyReader(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), responseBody)
	}

	_, err = c.checkTaskStatus(response.Header.Get("Location"))
	if err != nil {
		return err
	}

	return nil
}

func (c Client) Stop(deployment, job string, index int) error {
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/deployments/%s/jobs/%s/%d?state=stop&soft=true", c.config.URL, deployment, job, index), bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	request.Header.Set("Content-Type", "text/yaml")
	response, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusFound {
		responseBody, err := bodyReader(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), responseBody)
	}

	_, err = c.checkTaskStatus(response.Header.Get("Location"))
	if err != nil {
		return err
	}

	return nil
}

func (c Client) GetTaskOutput(taskId int) ([]TaskOutput, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/tasks/%d/output?type=event", c.config.URL, taskId), nil)
	if err != nil {
		return []TaskOutput{}, err
	}
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return []TaskOutput{}, err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return []TaskOutput{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return []TaskOutput{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	body = bytes.TrimSpace(body)
	parts := bytes.Split(body, []byte("\n"))

	var taskOutputs []TaskOutput
	for _, part := range parts {
		var taskOutput TaskOutput
		err = json.Unmarshal(part, &taskOutput)
		if err != nil {
			return []TaskOutput{}, err
		}

		taskOutputs = append(taskOutputs, taskOutput)
	}

	return taskOutputs, nil
}

func (c Client) checkTaskStatus(location string) (int, error) {
	for {
		task, err := c.checkTask(location)
		if err != nil {
			return 0, err
		}

		switch task.State {
		case "done":
			return task.Id, nil
		case "error":
			taskOutputs, err := c.GetTaskOutput(task.Id)
			if err != nil {
				return task.Id, fmt.Errorf("failed to get full bosh task event log, bosh task failed with an error status %q", task.Result)
			}
			errorMessage := taskOutputs[len(taskOutputs)-1].Error.Message
			return task.Id, fmt.Errorf("bosh task failed with an error status %q", errorMessage)
		case "errored":
			taskOutputs, err := c.GetTaskOutput(task.Id)
			if err != nil {
				return task.Id, fmt.Errorf("failed to get full bosh task event log, bosh task failed with an errored status %q", task.Result)
			}
			errorMessage := taskOutputs[len(taskOutputs)-1].Error.Message
			return task.Id, fmt.Errorf("bosh task failed with an errored status %q", errorMessage)
		case "cancelled":
			return task.Id, errors.New("bosh task was cancelled")
		default:
			time.Sleep(c.config.TaskPollingInterval)
		}
	}
}

func (c Client) checkTask(location string) (Task, error) {
	location, err := c.rewriteURL(location)
	if err != nil {
		return Task{}, err
	}

	var task Task
	request, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return task, err
	}
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return task, err
	}

	err = json.NewDecoder(response.Body).Decode(&task)
	if err != nil {
		return task, err
	}

	return task, nil
}

func (c Client) rewriteURL(uri string) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	parsedURL.Scheme = ""
	parsedURL.Host = ""

	return c.config.URL + parsedURL.String(), nil
}

