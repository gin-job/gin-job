package handler

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/gin-job/gin-job/config"

	"github.com/gin-gonic/gin"
)

type UIRoutes struct {
	authConfig *config.GinJobAuth
}

func NewUIRoutes(authConfig *config.GinJobAuth) *UIRoutes {
	return &UIRoutes{
		authConfig: authConfig,
	}
}

func (h *UIRoutes) RegisterRoutes(r *gin.Engine) {
	// 登录相关路由（不需要认证）
	r.GET("/ui/login", h.loginPage)
	r.POST("/ui/login", h.handleLogin)
	r.POST("/ui/logout", h.handleLogout)

	// 任务管理页面（需要认证）
	r.GET("/ui/jobs", h.checkAuth, h.jobsPage)
}

// checkAuth 检查登录状态的中间件
func (h *UIRoutes) checkAuth(c *gin.Context) {
	// 从 cookie 中获取 token
	token, err := c.Cookie("job_scheduler_token")
	if err != nil || token == "" {
		c.Redirect(http.StatusFound, "/ui/login")
		c.Abort()
		return
	}

	// 验证 token（这里使用简单的 MD5 验证，实际项目中可以使用更安全的方式）
	expectedToken := h.generateToken()
	if token != expectedToken {
		c.Redirect(http.StatusFound, "/ui/login")
		c.Abort()
		return
	}

	c.Next()
}

// loginPage 显示登录页面
func (h *UIRoutes) loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

// handleLogin 处理登录请求
func (h *UIRoutes) handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 从配置文件读取用户名和密码
	configUsername := h.authConfig.Username
	configPassword := h.authConfig.Password

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "用户名和密码不能为空",
		})
		return
	}

	if username != configUsername || password != configPassword {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "用户名或密码错误",
		})
		return
	}

	// 登录成功，设置 cookie
	token := h.generateToken()
	c.SetCookie("job_scheduler_token", token, 3600*24, "/", "", false, true) // 24小时有效

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
	})
}

// handleLogout 处理退出登录
func (h *UIRoutes) handleLogout(c *gin.Context) {
	c.SetCookie("job_scheduler_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "退出成功",
	})
}

// jobsPage 任务管理页面
func (h *UIRoutes) jobsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "jobs.html", gin.H{})
}

// generateToken 生成简单的 token（基于配置的用户名和密码）
func (h *UIRoutes) generateToken() string {
	username := h.authConfig.Username
	password := h.authConfig.Password
	// 使用用户名+密码+固定盐值生成 token
	data := fmt.Sprintf("%s:%s:job_scheduler_2024", username, password)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}
