// 调度器核心配置
package scheduler

import (
	"context"
	"errors"
	"sync"

	cron "github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/gin-job/gin-job/job"
	"github.com/gin-job/gin-job/model"

	"gorm.io/gorm"
)

// Scheduler 定时任务调度器核心结构体
type Scheduler struct {
	c       *cron.Cron              // robfig/cron 实例，实际负责定时调度,底层负责解析 Cron 表达式、管理任务执行时机
	log     *zap.Logger             //日志
	db      *gorm.DB                // 数据库实例
	mu      sync.Mutex              //互斥锁
	entries map[string]cron.EntryID //任务名与调度项 ID 的映射表，用于快速查询/移除已注册任务
}

// New 创建调度器实例
func New(log *zap.Logger, db *gorm.DB) *Scheduler {
	c := cron.New(cron.WithLogger(cron.PrintfLogger(zap.NewStdLog(log))))
	return &Scheduler{
		c:       c,
		log:     log,
		db:      db,
		entries: make(map[string]cron.EntryID),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() { s.c.Start() }

// Stop 关闭调度器
func (s *Scheduler) Stop(ctx context.Context) {
	stopCtx := s.c.Stop()
	//等待当前任务完成后再退出
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	}
}

// SyncFromDB 启动时同步数据库里的任务（启用的才注册）
func (s *Scheduler) SyncFromDB() error {
	if err := s.db.AutoMigrate(&model.SysJobScheduleModel{}, &model.SysJobInstanceModel{}); err != nil {
		return err
	}
	var jobs []model.SysJobScheduleModel
	if err := s.db.Order("id asc").Find(&jobs).Error; err != nil {
		return err
	}
	for _, j := range jobs {
		if j.Enabled {
			if err := s.upsertSchedule(j); err != nil {
				fields := []zap.Field{zap.String("job", j.Name), zap.Error(err)}
				s.log.Warn("failed to schedule job", fields...)
			}
		}
	}
	return nil
}

// Upsert 更新或创建任务，持久化,并根据 enabled 注册/移除,启用则注册，禁用则移除
func (s *Scheduler) Upsert(rec *model.SysJobScheduleModel) error {
	if rec.Name == "" {
		return errors.New("job name required")
	}
	if rec.HandlerName == "" {
		return errors.New("job handler_name required")
	}
	if rec.Spec == "" {
		return errors.New("job spec required")
	}

	// 验证 handler 是否存在
	if _, err := job.Get(rec.HandlerName); err != nil {
		return errors.New("job handler not found: " + rec.HandlerName)
	}

	// 保存到DB
	var exist model.SysJobScheduleModel
	if err := s.db.Where("name = ?", rec.Name).First(&exist).Error; err == nil {
		// 更新已存在的记录
		exist.HandlerName = rec.HandlerName
		exist.Spec = rec.Spec
		exist.Description = rec.Description
		exist.Enabled = rec.Enabled
		// 确保 Name 也被更新（防止之前记录中 name 为空的情况）
		exist.Name = rec.Name
		if err := s.db.Select("Name", "HandlerName", "Spec", "Enabled", "Description").Save(&exist).Error; err != nil {
			return err
		}
		rec.BaseModel.Id = exist.BaseModel.Id
	} else {
		// 创建新记录
		if err := s.db.Select("Name", "HandlerName", "Spec", "Enabled", "Description").Create(rec).Error; err != nil {
			return err
		}
	}
	// 应用到调度
	if rec.Enabled {
		return s.upsertSchedule(*rec)
	}
	s.Disable(rec.Name)
	return nil
}

// Enable 开启并注册
func (s *Scheduler) Enable(name string) error {
	var rec model.SysJobScheduleModel
	if err := s.db.Where("name = ?", name).First(&rec).Error; err != nil {
		return err
	}
	rec.Enabled = true
	if err := s.db.Save(&rec).Error; err != nil {
		return err
	}
	return s.upsertSchedule(rec)
}

// Disable 关闭并移除调度
func (s *Scheduler) Disable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.entries[name]; ok {
		s.c.Remove(id)
		delete(s.entries, name)
	}
	return s.db.Model(&model.SysJobScheduleModel{}).Where("name = ?", name).Update("enabled", false).Error
}

// UpdateSpec 修改频率cron
func (s *Scheduler) UpdateSpec(name, spec string) error {
	var rec model.SysJobScheduleModel
	if err := s.db.Where("name = ?", name).First(&rec).Error; err != nil {
		return err
	}
	rec.Spec = spec
	if err := s.db.Save(&rec).Error; err != nil {
		return err
	}
	// 若任务启用，立即同步到调度器（生效新的执行频率）
	if rec.Enabled {
		return s.upsertSchedule(rec)
	}
	return nil
}

// Trigger 立即触发一次（不修改下次调度）
func (s *Scheduler) Trigger(name string) error {
	var rec model.SysJobScheduleModel
	if err := s.db.Where("name = ?", name).First(&rec).Error; err != nil {
		return errors.New("job not found: " + name)
	}

	// 获取任务实现
	jobImpl, err := job.Get(rec.HandlerName)
	if err != nil {
		return err
	}

	// 使用包装器执行任务
	wrapper := newJobWrapper(rec.Name, uint(rec.BaseModel.Id), jobImpl, s.log, s.db)
	go wrapper.Run()

	return nil
}

// GetAvailableHandlers 获取所有可用的任务处理器
func (s *Scheduler) GetAvailableHandlers() map[string]string {
	return job.List()
}

// Delete 删除任务（从调度器中移除并从数据库删除）
func (s *Scheduler) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 先从调度器中移除（如果已注册）
	if id, ok := s.entries[name]; ok {
		s.c.Remove(id)
		delete(s.entries, name)
		fields := []zap.Field{zap.String("job", name)}
		s.log.Info("removed job from scheduler", fields...)
	}

	// 从数据库删除任务
	result := s.db.Where("name = ?", name).Delete(&model.SysJobScheduleModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("job not found: " + name)
	}

	fields := []zap.Field{zap.String("job", name)}
	s.log.Info("deleted job", fields...)
	return nil
}

// 内部：根据记录注册或更新调度项，将任务注册到 cron 调度器
func (s *Scheduler) upsertSchedule(rec model.SysJobScheduleModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 若任务已注册，先移除旧的调度项（避免重复注册）
	if id, ok := s.entries[rec.Name]; ok {
		s.c.Remove(id)
		delete(s.entries, rec.Name)
		fields := []zap.Field{zap.String("job", rec.Name)}
		s.log.Info("removed existing schedule", fields...)
	}

	// 从注册表获取任务实现
	jobImpl, err := job.Get(rec.HandlerName)
	if err != nil {
		fields := []zap.Field{zap.String("job", rec.Name), zap.String("handler", rec.HandlerName), zap.Error(err)}
		s.log.Warn("unknown job handler", fields...)
		return errors.New("job handler not found: " + rec.HandlerName)
	}

	// 创建包装器，用于记录执行历史
	wrapper := newJobWrapper(rec.Name, uint(rec.BaseModel.Id), jobImpl, s.log, s.db)

	fields := []zap.Field{zap.String("job", rec.Name), zap.String("handler", rec.HandlerName), zap.String("spec", rec.Spec)}
	s.log.Info("adding job to scheduler", fields...)
	// 将任务和 Cron 表达式添加到调度器，返回调度项 ID
	newID, err := s.c.AddJob(rec.Spec, wrapper)
	if err != nil {
		fields := []zap.Field{zap.String("job", rec.Name), zap.String("spec", rec.Spec), zap.Error(err)}
		s.log.Error("failed to add job to scheduler", fields...)
		return err
	}
	// 保存任务名与调度项 ID 的映射，用于修改/移除
	s.entries[rec.Name] = newID
	successFields := []zap.Field{zap.String("job", rec.Name), zap.String("spec", rec.Spec), zap.Int("entry_id", int(newID))}
	s.log.Info("job scheduled successfully", successFields...)
	return nil
}
