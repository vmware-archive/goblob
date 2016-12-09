package cc

type CloudController struct {
	Status string
}

func NewCloudController() CloudController {
	return CloudController{}
}

func (c CloudController) Start() error {
	c.Status = "started"
	return nil
}

func (c CloudController) Stop() error {
	c.Status = "stopped"
	return nil
}
