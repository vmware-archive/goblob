package fakes

type FakeBoshClient struct {

}

func NewFakeBoshClient() FakeBoshClient {
	return FakeBoshClient{}
}

func (c FakeBoshClient) Start(deployment, job string, index int) error {
	return nil
}

func (c FakeBoshClient) Stop(deployment, job string, index int) error {
	return nil
}