package cc

import "github.com/c0-ops/goblob/bosh"

type BOSHClient interface {
	Start(deployment, job string, index int) error
	Stop(deployment, job string, index int) error
}

type CloudController struct {
	deployment string
	status     string
	jobs       []bosh.VM

	client BOSHClient
}

func NewCloudController(client BOSHClient, deployment string, jobs []bosh.VM) *CloudController {
	return &CloudController{
		client:     client,
		deployment: deployment,
		jobs:       jobs,
	}
}

func (c *CloudController) Start() error {
	c.status = "started"
	for _, job := range c.jobs {
		err := c.client.Start(c.deployment, job.JobName, job.Index)
		if err != nil {
			c.status = "errored"
			return err
		}
	}

	return nil
}

func (c *CloudController) Stop() error {
	c.status = "stopped"
	for _, job := range c.jobs {
		err := c.client.Stop(c.deployment, job.JobName, job.Index)
		if err != nil {
			c.status = "errored"
			return err
		}
	}

	return nil
}

func (c *CloudController) GetStatus() string {
	return c.status
}
