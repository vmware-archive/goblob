package fakes

type FakeCmdExtractor struct {
	ExtractPath     string
}

func NewFakeCmdExtractor() FakeCmdExtractor {
	return FakeCmdExtractor{
		ExtractPath: "",
	}
}

func (e FakeCmdExtractor) Extract(path string) (string, error) {

	return e.ExtractPath, nil
}

func (e FakeCmdExtractor) CleanUp(path string) error {
	return nil
}