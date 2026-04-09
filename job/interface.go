package job

import "context"

// Job 任务接口，所有任务实现都需要实现这个接口
type Job interface {
	// Name 返回任务的唯一标识名称
	Name() string

	// Run 执行任务
	Run(ctx context.Context) error

	// Description 返回任务的描述信息
	Description() string
}
