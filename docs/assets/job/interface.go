package job

import "context"

// Job Interface
type Job interface {
	// Return Job Name
	Name() string

	// Run Job
	Run(ctx context.Context) error

	// Return Job Description
	Description() string
}
