# JobSchedule 和 JobInstance 详解

## 核心概念

### 1. JobSchedule (任务调度)
`JobSchedule` 是定时任务的配置信息，对应数据库表 `tb_sys_job_schedule`。它包含以下主要字段：

| 字段名 | 类型 | 描述 |
|-------|------|------|
| Name | string | 任务名称（唯一索引） |
| HandlerName | string | 处理器名称（对应代码中定义的任务） |
| Spec | string | Cron 表达式（定时执行规则） |
| Enabled | bool | 是否启用 |
| Description | string | 任务描述 |
| LastRunAt | *time.Time | 上次执行时间 |
| Status | string | 任务状态 |
| LastError | string | 上次执行错误信息 |

### 2. JobInstance (任务实例)
`JobInstance` 是任务的执行实例，对应数据库表 `tb_sys_job_instance`。它记录了每次任务执行的详细信息：

| 字段名 | 类型 | 描述 |
|-------|------|------|
| JobName | string | 任务名称 |
| JobID | uint | 任务 ID |
| Status | string | 执行状态（running, success, failed） |
| StartedAt | time.Time | 开始执行时间 |
| FinishedAt | *time.Time | 结束执行时间 |
| DurationMs | int64 | 执行时长（毫秒） |
| Error | string | 执行错误信息 |
| LogContent | string | 执行日志内容 |

## 如何在页面上开启并触发执行任务

### 步骤 1：在代码中定义任务
首先，你需要实现 `Job` 接口来定义一个任务：

```go
package jobs

import (
    "context"
    "fmt"
)

// ExampleJob 示例任务
type ExampleJob struct{}

// Name 返回任务名称
func (j *ExampleJob) Name() string {
    return "example_job"
}

// Run 执行任务逻辑
func (j *ExampleJob) Run(ctx context.Context) error {
    fmt.Println("Example job is running...")
    return nil
}

// Description 返回任务描述
func (j *ExampleJob) Description() string {
    return "示例任务"
}
```

### 步骤 2：注册任务
在应用启动时，将任务注册到调度器：

```go
// register job
jobList := []job.Job{
    &jobs.ExampleJob{},
}

// init router
r := router.NewGinJobRouter(nil)
r.SetJobList(jobList)
r.Start()
```

### 步骤 3：在 Web 界面管理任务
应用启动后，你可以通过 Web 界面来管理任务：

![UI 界面](assets/images/ui.jpg)

1. **新建任务**：
   - 访问 Web 界面（通常是 http://localhost:8080）
   - 点击 "新建任务" 按钮
   - 填写任务名称、选择处理器（ExampleJob）、设置 Cron 表达式、描述等
   - 点击 "保存" 按钮

2. **启用任务**：
   - 在任务列表中找到刚新建的任务
   - 点击 "启用" 按钮，任务会按照设定的 Cron 表达式自动执行

3. **手动触发任务**：
   - 在任务列表中找到任务
   - 点击 "立即执行" 按钮，任务会立即执行一次，不影响定时执行规则

4. **查看执行历史**：
   - 在任务详情页面，可以查看该任务的所有执行实例
   - 点击具体的执行实例，可以查看详细的执行日志和结果

### 核心 API 接口

| 接口 | 方法 | 描述 |
|------|------|------|
| /jobs | GET | 获取任务列表 |
| /jobs | POST | 创建新任务 |
| /jobs/:name | GET | 获取任务详情 |
| /jobs/:name | PUT | 修改任务 |
| /jobs/:name | DELETE | 删除任务 |
| /jobs/:name/enable | POST | 启用任务 |
| /jobs/:name/disable | POST | 停用任务 |
| /jobs/:name/trigger | POST | 立即触发任务 |
| /jobs/:name/runs | GET | 获取任务执行历史 |
| /jobs/:name/runs/:id | GET | 获取任务执行详情 |
| /jobs/handlers | GET | 获取可用的任务处理器列表 |

## 工作流程

1. **任务定义**：在代码中实现 `Job` 接口
2. **任务注册**：应用启动时将任务注册到调度器
3. **任务配置**：通过 Web 界面创建任务调度（JobSchedule）
4. **任务执行**：
   - 定时执行：根据 Cron 表达式自动执行
   - 手动执行：通过 "立即执行" 按钮触发
5. **执行记录**：每次执行都会创建一个 JobInstance 记录
6. **状态管理**：可以通过启用/禁用接口控制任务状态

通过以上流程，你可以在代码中定义任务，并在 Web 界面上方便地管理和触发任务执行。

## 配置说明

GinJob 提供了以下配置项，您可以根据需要进行调整：

### 配置结构体

```go
type GinJobConfig struct {
    TemplatePath string    // 模板文件路径
    Auth         GinJobAuth // 认证信息
    Port         string     // 服务端口
    Gorm         GinJobGorm // 数据库配置
}

type GinJobAuth struct {
    Username string // 登录用户名
    Password string // 登录密码
}

type GinJobGorm struct {
    DSN    string      // 数据库连接字符串
    Config *gorm.Config // GORM 配置
}
```

### 默认配置

GinJob 提供了默认配置，您可以直接使用：

```go
func DefaultConfig() *GinJobConfig {
    templatePath := os.Getenv("TEMPLATE_PATH")
    if templatePath == "" {
        templatePath = "../../templates/*"
    }
    gormConfig := &gorm.Config{}
    dsn := "root:gin-job@tcp(localhost:3306)/gin_job?charset=utf8mb4&parseTime=True&loc=Local"
    return &GinJobConfig{
        Port: ":8080",
        Gorm: GinJobGorm{
            DSN:    dsn,
            Config: gormConfig,
        },
        TemplatePath: templatePath,
        Auth: GinJobAuth{
            Username: "admin",
            Password: "gin-job",
        },
    }
}
```

### 配置项说明

| 配置项 | 类型 | 默认值 | 说明 |
|-------|------|-------|------|
| TemplatePath | string | ../../templates/* | 模板文件路径，可通过环境变量 TEMPLATE_PATH 覆盖 |
| Auth.Username | string | admin | 登录用户名 |
| Auth.Password | string | gin-job | 登录密码 |
| Port | string | :8080 | 服务端口 |
| Gorm.DSN | string | root:gin-job@tcp(localhost:3306)/gin_job?charset=utf8mb4&parseTime=True&loc=Local | 数据库连接字符串 |
| Gorm.Config | *gorm.Config | &gorm.Config{} | GORM 配置对象 |

### 如何使用配置

在初始化 GinJob 路由器时，您可以传入自定义配置：

```go
// 创建自定义配置
customConfig := &config.GinJobConfig{
    Port: ":9090",
    Auth: config.GinJobAuth{
        Username: "custom",
        Password: "custom-password",
    },
    Gorm: config.GinJobGorm{
        DSN: "user:pass@tcp(localhost:3306)/custom_db?charset=utf8mb4&parseTime=True&loc=Local",
    },
}

// 初始化路由器并传入配置
r := router.NewGinJobRouter(customConfig)
r.SetJobList(jobList)
r.Start()
```

如果您不传入配置，GinJob 会使用默认配置：

```go
// 使用默认配置
r := router.NewGinJobRouter(nil)
r.SetJobList(jobList)
r.Start()
```