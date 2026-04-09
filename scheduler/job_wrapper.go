package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gin-job/gin-job/job"
	"github.com/gin-job/gin-job/model"

	"gorm.io/gorm"
)

// jobWrapper 包装任务执行，记录执行历史和日志
type jobWrapper struct {
	jobName string
	jobID   uint
	jobImpl job.Job
	log     *zap.Logger
	db      *gorm.DB
}

func newJobWrapper(jobName string, jobID uint, jobImpl job.Job, log *zap.Logger, db *gorm.DB) *jobWrapper {
	return &jobWrapper{
		jobName: jobName,
		jobID:   jobID,
		jobImpl: jobImpl,
		log:     log,
		db:      db,
	}
}

func (w *jobWrapper) Run() {
	ctx := context.Background()
	startTime := time.Now()

	// 创建 jobInstance 记录
	jobInstance := &model.SysJobInstanceModel{
		JobName:   w.jobName,
		JobID:     w.jobID,
		Status:    "running",
		StartedAt: startTime,
	}

	if err := w.db.Create(jobInstance).Error; err != nil {
		w.log.Error("failed to create job run", zap.Error(err))
		return
	}

	instanceID := jobInstance.BaseModel.Id

	// 捕获日志
	logEntries := []string{}
	logEntries = append(logEntries, fmt.Sprintf("[%s] 任务开始执行: %s", startTime.Format("2006-01-02 15:04:05"), w.jobName))

	// 执行任务
	jobErr := w.jobImpl.Run(ctx)

	finishTime := time.Now()
	duration := finishTime.Sub(startTime)

	// 更新 jobInstance 记录
	jobInstance.FinishedAt = &finishTime
	jobInstance.DurationMs = duration.Milliseconds()

	var logContent string
	if jobErr != nil {
		jobInstance.Status = "failed"
		jobInstance.Error = jobErr.Error()
		logEntries = append(logEntries, fmt.Sprintf("[%s] 任务执行失败: %v", finishTime.Format("2006-01-02 15:04:05"), jobErr))
		fields := []zap.Field{zap.String("job", w.jobName), zap.Error(jobErr)}
		w.log.Error("job execution failed", fields...)
	} else {
		jobInstance.Status = "success"
		logEntries = append(logEntries, fmt.Sprintf("[%s] 任务执行成功，耗时: %v", finishTime.Format("2006-01-02 15:04:05"), duration))
		fields := []zap.Field{zap.String("job", w.jobName), zap.Duration("duration", duration)}
		w.log.Info("job execution success", fields...)
	}

	logContent = strings.Join(logEntries, "\n")
	jobInstance.LogContent = logContent

	// 保存 jobInstance
	if err := w.db.Save(jobInstance).Error; err != nil {
		w.log.Error("failed to save job run", zap.Error(err))
	}

	// 更新 Job 表的状态
	var jobRecord model.SysJobScheduleModel
	if err := w.db.Where("id = ?", w.jobID).First(&jobRecord).Error; err == nil {
		jobRecord.LastRunAt = &finishTime
		if jobErr != nil {
			jobRecord.Status = "error"
			jobRecord.LastError = jobErr.Error()
		} else {
			jobRecord.Status = "ok"
			jobRecord.LastError = ""
		}
		_ = w.db.Save(&jobRecord).Error
	}

	fields := []zap.Field{
		zap.Int64("instance_id", instanceID),
		zap.String("job", w.jobName),
		zap.String("status", jobInstance.Status),
		zap.Duration("duration", duration),
	}
	w.log.Info("job run completed", fields...)
}
