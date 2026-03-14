package git

// Client is the interface that the cmd layer uses for all git/gh operations.
// A real implementation wraps the package-level functions; tests inject a mock.
type Client interface {
	Tags() ([]string, error)
	TagExists(tag string) bool
	CreateTag(tag, message string) error
	PushTag(tag string) error
	DeleteTag(tag string) error
	CreateRelease(tag string, opts ReleaseOptions) error
}

// NewClient returns a Client backed by the real git and gh binaries.
func NewClient() Client { return &realClient{} }

type realClient struct{}

func (c *realClient) Tags() ([]string, error)          { return Tags() }
func (c *realClient) TagExists(tag string) bool         { return TagExists(tag) }
func (c *realClient) CreateTag(tag, msg string) error   { return CreateTag(tag, msg) }
func (c *realClient) PushTag(tag string) error          { return PushTag(tag) }
func (c *realClient) DeleteTag(tag string) error        { return DeleteTag(tag) }
func (c *realClient) CreateRelease(tag string, opts ReleaseOptions) error {
	return CreateRelease(tag, opts)
}
