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

// Scheduler core struct
type Scheduler struct {
	c       *cron.Cron              // robfig/cron instance
	log     *zap.Logger             // logger
	db      *gorm.DB                // db instance
	mu      sync.Mutex              // mutex
	entries map[string]cron.EntryID // job name to entry ID map
}

// New scheduler instance
func New(log *zap.Logger, db *gorm.DB) *Scheduler {
	c := cron.New(cron.WithLogger(cron.PrintfLogger(zap.NewStdLog(log))))
	return &Scheduler{
		c:       c,
		log:     log,
		db:      db,
		entries: make(map[string]cron.EntryID),
	}
}

// Start scheduler
func (s *Scheduler) Start() { s.c.Start() }

// Stop scheduler
func (s *Scheduler) Stop(ctx context.Context) {
	stopCtx := s.c.Stop()
	//等待当前任务完成后再退出
	select {
	case <-stopCtx.Done():
	case <-ctx.Done():
	}
}

// SyncFromDB sync jobs from db
// Only enabled jobs are registered.
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

// Upsert
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

	if _, err := job.Get(rec.HandlerName); err != nil {
		return errors.New("job handler not found: " + rec.HandlerName)
	}

	// Save to DB
	var exist model.SysJobScheduleModel
	if err := s.db.Where("name = ?", rec.Name).First(&exist).Error; err == nil {
		// Update existing record
		exist.HandlerName = rec.HandlerName
		exist.Spec = rec.Spec
		exist.Description = rec.Description
		exist.Enabled = rec.Enabled
		// Ensure Name is updated as well
		exist.Name = rec.Name
		if err := s.db.Select("Name", "HandlerName", "Spec", "Enabled", "Description").Save(&exist).Error; err != nil {
			return err
		}
		rec.BaseModel.Id = exist.BaseModel.Id
	} else {
		// Create new record
		if err := s.db.Select("Name", "HandlerName", "Spec", "Enabled", "Description").Create(rec).Error; err != nil {
			return err
		}
	}
	// Apply to scheduler
	if rec.Enabled {
		return s.upsertSchedule(*rec)
	}
	s.Disable(rec.Name)
	return nil
}

// Enable job in scheduler
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

// Disable job in scheduler
func (s *Scheduler) Disable(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.entries[name]; ok {
		s.c.Remove(id)
		delete(s.entries, name)
	}
	return s.db.Model(&model.SysJobScheduleModel{}).Where("name = ?", name).Update("enabled", false).Error
}

// Update job spec
func (s *Scheduler) UpdateSpec(name, spec string) error {
	var rec model.SysJobScheduleModel
	if err := s.db.Where("name = ?", name).First(&rec).Error; err != nil {
		return err
	}
	rec.Spec = spec
	if err := s.db.Save(&rec).Error; err != nil {
		return err
	}
	// Apply to scheduler
	if rec.Enabled {
		return s.upsertSchedule(rec)
	}
	return nil
}

// Trigger job once
func (s *Scheduler) Trigger(name string) error {
	var rec model.SysJobScheduleModel
	if err := s.db.Where("name = ?", name).First(&rec).Error; err != nil {
		return errors.New("job not found: " + name)
	}

	// Get job implementation
	jobImpl, err := job.Get(rec.HandlerName)
	if err != nil {
		return err
	}

	// Run job once
	wrapper := newJobWrapper(rec.Name, uint(rec.BaseModel.Id), jobImpl, s.log, s.db)
	go wrapper.Run()

	return nil
}

// all available job handlers
func (s *Scheduler) GetAvailableHandlers() map[string]string {
	return job.List()
}

// Delete job from scheduler
func (s *Scheduler) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// remove job from scheduler
	if id, ok := s.entries[name]; ok {
		s.c.Remove(id)
		delete(s.entries, name)
		fields := []zap.Field{zap.String("job", name)}
		s.log.Info("removed job from scheduler", fields...)
	}

	// delete job from db
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

// Upsert job schedule
func (s *Scheduler) upsertSchedule(rec model.SysJobScheduleModel) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// remove existing schedule
	if id, ok := s.entries[rec.Name]; ok {
		s.c.Remove(id)
		delete(s.entries, rec.Name)
		fields := []zap.Field{zap.String("job", rec.Name)}
		s.log.Info("removed existing schedule", fields...)
	}

	// Get job implementation
	jobImpl, err := job.Get(rec.HandlerName)
	if err != nil {
		fields := []zap.Field{zap.String("job", rec.Name), zap.String("handler", rec.HandlerName), zap.Error(err)}
		s.log.Warn("unknown job handler", fields...)
		return errors.New("job handler not found: " + rec.HandlerName)
	}

	// Create job wrapper
	wrapper := newJobWrapper(rec.Name, uint(rec.BaseModel.Id), jobImpl, s.log, s.db)

	fields := []zap.Field{zap.String("job", rec.Name), zap.String("handler", rec.HandlerName), zap.String("spec", rec.Spec)}
	s.log.Info("adding job to scheduler", fields...)
	// Add job to scheduler
	newID, err := s.c.AddJob(rec.Spec, wrapper)
	if err != nil {
		fields := []zap.Field{zap.String("job", rec.Name), zap.String("spec", rec.Spec), zap.Error(err)}
		s.log.Error("failed to add job to scheduler", fields...)
		return err
	}
	// Save job schedule
	s.entries[rec.Name] = newID
	successFields := []zap.Field{zap.String("job", rec.Name), zap.String("spec", rec.Spec), zap.Int("entry_id", int(newID))}
	s.log.Info("job scheduled successfully", successFields...)
	return nil
}
