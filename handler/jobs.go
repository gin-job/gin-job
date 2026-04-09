package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gin-job/gin-job/model"
	"github.com/gin-job/gin-job/scheduler"

	"gorm.io/gorm"
)

type JobHandler struct {
	sch *scheduler.Scheduler // 调度器实例，关联核心业务逻辑
	log *zap.Logger          // 日志实例，记录接口请求与错误
	db  *gorm.DB             // 数据库实例
}

func NewJobHandler(s *scheduler.Scheduler, log *zap.Logger, db *gorm.DB) *JobHandler {
	return &JobHandler{sch: s, log: log, db: db}
}

func (h *JobHandler) RegisterRoutes(r *gin.Engine) {
	grp := r.Group("/jobs")

	grp.GET("", h.list)                  //获取任务列表
	grp.GET("/handlers", h.listHandlers) //获取可用的任务处理器列表

	grp.POST("", h.create) //创建新任务

	grp.GET("/:name/runs/:id", h.getRun)  //获取任务执行详情
	grp.GET("/:name/runs", h.listRuns)    //获取任务执行历史
	grp.POST("/:name/enable", h.enable)   //启用任务
	grp.POST("/:name/disable", h.disable) //停用任务
	grp.POST("/:name/trigger", h.trigger) //立即触发任务

	grp.GET("/:name", h.detail)    //获取任务详情
	grp.PUT("/:name", h.update)    //修改任务
	grp.DELETE("/:name", h.delete) //删除任务
}

// 获取任务列表
func (h *JobHandler) list(c *gin.Context) {
	var jobs []model.SysJobScheduleModel
	if err := h.db.Order("id asc").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// 获取可用的任务处理器列表
func (h *JobHandler) listHandlers(c *gin.Context) {
	handlers := h.sch.GetAvailableHandlers()
	c.JSON(http.StatusOK, handlers)
}

// 获取任务详情
func (h *JobHandler) detail(c *gin.Context) {
	name := c.Param("name")
	var job model.SysJobScheduleModel
	if err := h.db.Where("name = ?", name).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// 获取任务执行历史
func (h *JobHandler) listRuns(c *gin.Context) {
	name := c.Param("name")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}

	// 验证任务是否存在
	var job model.SysJobScheduleModel
	if err := h.db.Where("name = ?", name).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	var instance []model.SysJobInstanceModel
	if err := h.db.
		Where("job_name = ?", name).
		Order("started_at desc").
		Limit(limit).
		Find(&instance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, instance)
}

// 获取任务执行详情
func (h *JobHandler) getRun(c *gin.Context) {
	name := c.Param("name")
	runIDStr := c.Param("id")
	runID, err := strconv.ParseUint(runIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid run id"})
		return
	}

	var run model.SysJobInstanceModel
	if err := h.db.
		Where("id = ? AND job_name = ?", runID, name).
		First(&run).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "run not found"})
		return
	}
	c.JSON(http.StatusOK, run)
}

// 创建新任务
func (h *JobHandler) create(c *gin.Context) {
	var req model.SysJobScheduleModel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 验证必填字段
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "任务名称不能为空"})
		return
	}
	// 调用调度器的 Upsert 方法创建任务
	if err := h.sch.Upsert(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// 修改任务
func (h *JobHandler) update(c *gin.Context) {
	name := c.Param("name")
	var req struct {
		HandlerName string `json:"handler_name"`
		Spec        string `json:"spec"`
		Enabled     *bool  `json:"enabled"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var rec model.SysJobScheduleModel
	if err := h.db.Where("name = ?", name).First(&rec).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if req.HandlerName != "" {
		rec.HandlerName = req.HandlerName
		fields := []zap.Field{zap.String("name", name), zap.String("handler", req.HandlerName)}
		h.log.Info("updating job handler", fields...)
	}
	if req.Spec != "" {
		rec.Spec = req.Spec
		fields := []zap.Field{zap.String("name", name), zap.String("spec", req.Spec)}
		h.log.Info("updating job spec", fields...)
	}
	if req.Enabled != nil {
		rec.Enabled = *req.Enabled
		fields := []zap.Field{zap.String("name", name), zap.Bool("enabled", *req.Enabled)}
		h.log.Info("updating job enabled", fields...)
	}
	if req.Description != "" {
		rec.Description = req.Description
	}
	if err := h.sch.Upsert(&rec); err != nil {
		fields := []zap.Field{zap.String("name", name), zap.Error(err)}
		h.log.Error("failed to upsert job", fields...)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rec)
}

// 启用任务
func (h *JobHandler) enable(c *gin.Context) {
	name := c.Param("name")
	if err := h.sch.Enable(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// 停用任务
func (h *JobHandler) disable(c *gin.Context) {
	name := c.Param("name")
	if err := h.sch.Disable(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// 立即触发任务
func (h *JobHandler) trigger(c *gin.Context) {
	name := c.Param("name")
	if err := h.sch.Trigger(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// 删除任务
func (h *JobHandler) delete(c *gin.Context) {
	name := c.Param("name")
	fields := []zap.Field{zap.String("name", name), zap.String("path", c.Request.URL.Path)}
	h.log.Info("delete job request", fields...)
	if err := h.sch.Delete(name); err != nil {
		fields := []zap.Field{zap.String("name", name), zap.Error(err)}
		h.log.Warn("delete job failed", fields...)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fields = []zap.Field{zap.String("name", name)}
	h.log.Info("delete job success", fields...)
	c.Status(http.StatusOK)
}
