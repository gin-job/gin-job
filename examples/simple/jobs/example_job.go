package jobs

import (
	"context"
)

type ExampleJob struct{}

func (j ExampleJob) Name() string {
	return "example_job"
}

func (j ExampleJob) Description() string {
	return "示例任务"
}

func (j ExampleJob) Run(ctx context.Context) error {
	// 任务执行逻辑
	return nil
}
