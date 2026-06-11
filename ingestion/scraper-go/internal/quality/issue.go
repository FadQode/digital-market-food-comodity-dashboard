package quality

type Issue struct {
	ScrapeRunID  int64
	ProductRawID *int64
	Source       string
	Severity     string
	IssueCode    string
	Message      string
	FieldName    string
	RawValue     string
	Metadata     map[string]any
}
