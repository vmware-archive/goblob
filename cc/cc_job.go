package cc

type BOSHClient interface {
	Start(deployment, job string, index int) error
	Stop(deployment, job string, index int) error
}

type CCJob struct {
	Name   string
	Deployment string
	Index int
}

type CloudController struct {
	status string
	jobs []CCJob

	client BOSHClient
}

func NewCloudController(client BOSHClient, jobs []CCJob) *CloudController {
	return &CloudController{
		client: client,
		jobs: jobs,
	}
}

func (c *CloudController) Start() error {
	c.status = "started"
	for _, job := range c.jobs {
		err := c.client.Start(job.Deployment, job.Name, job.Index)
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
		err := c.client.Stop(job.Deployment, job.Name, job.Index)
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
