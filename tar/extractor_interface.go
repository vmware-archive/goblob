package tar

type Extractor interface {
	Extract(string) (string, error)
	CleanUp(string) error
}
