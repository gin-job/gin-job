// Package ginjob provides a task scheduling management tool based on the Gin framework,
// with a user-friendly UI interface for managing scheduled tasks.
//
// Core Features:
// - Task scheduling with cron-like expressions
// - User-friendly web UI for task management
// - Task definition, execution, pause, resume, and deletion
// - Task history log query and monitoring
// - Integration with Gin framework for API endpoints
// - Detailed logging with Zap
// - Database persistence with GORM
//
// Core Concepts:
// - JobSchedule: Configuration information for scheduled tasks (stored in tb_sys_job_schedule table)
// - JobInstance: Execution instance of a task (stored in tb_sys_job_instance table)
//
// Project Structure:
// - config/: Configuration management
// - docs/: Documentation and assets
// - examples/: Example applications
// - handler/: HTTP handlers for API endpoints
// - job/: Job interface and registry
// - model/: Database models
// - router/: Gin router setup
// - scheduler/: Task scheduler implementation
// - templates/: UI templates
//
// Quick Start:
//
//  1. Initialize the local database:
//     docker run -it -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=gin-job -e MYSQL_DATABASE=gin_job mysql:8.0.42
//
//  2. Run the example application:
//     go run examples/simple/main.go
//
//  3. Open the UI interface:
//     http://localhost:8080/ui/login
//
// Example Usage:
//
//	import (
//	    "context"
//	    "fmt"
//	    "github.com/yourusername/gin-job/job"
//	    "github.com/yourusername/gin-job/router"
//	    "github.com/yourusername/gin-job/config"
//	    "go.uber.org/zap"
//	    "gorm.io/gorm"
//	)
//
//	// Define a task by implementing the Job interface
//	type ExampleJob struct{}
//
//	func (j *ExampleJob) Name() string {
//	    return "example_job"
//	}
//
//	func (j *ExampleJob) Run(ctx context.Context) error {
//	    fmt.Println("Example job is running...")
//	    return nil
//	}
//
//	func (j *ExampleJob) Description() string {
//	    return "Example task"
//	}
//
//	// Register tasks
//	jobList := []job.Job{
//	    &ExampleJob{},
//	}
//
//	// Initialize with default configuration
//	cfg := config.DefaultConfig()
//	// Or create custom configuration
//	// cfg := &config.GinJobConfig{
//	//     Port: ":9090",
//	//     Auth: config.GinJobAuth{
//	//         Username: "custom",
//	//         Password: "custom-password",
//	//     },
//	//     Gorm: config.GinJobGorm{
//	//         DSN: "user:pass@tcp(localhost:3306)/custom_db?charset=utf8mb4&parseTime=True&loc=Local",
//	//     },
//	// }
//
//	// Initialize router and start
//	r := router.NewGinJobRouter(cfg)
//	r.SetJobList(jobList)
//	r.Start()
//
// API Endpoints:
// - GET /jobs - Get task list
// - POST /jobs - Create new task
// - GET /jobs/:name - Get task details
// - PUT /jobs/:name - Modify task
// - DELETE /jobs/:name - Delete task
// - POST /jobs/:name/enable - Enable task
// - POST /jobs/:name/disable - Disable task
// - POST /jobs/:name/trigger - Trigger task immediately
// - GET /jobs/:name/runs - Get task execution history
// - GET /jobs/:name/runs/:id - Get task execution details
// - GET /jobs/handlers - Get list of available task handlers
//
// Configuration:
// GinJob provides a GinJobConfig structure for customization:
// - TemplatePath: Template file path
// - Auth: Authentication information (username/password)
// - Port: Service port
// - Gorm: Database configuration (DSN and GORM config)
//
// Default configuration uses:
// - Port: :8080
// - Username: admin
// - Password: gin-job
// - Database: root:gin-job@tcp(localhost:3306)/gin_job
// - TemplatePath: ../../templates/* (can be overridden by TEMPLATE_PATH environment variable)
package ginjob
