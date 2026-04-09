package main

import (
	"github.com/gin-job/gin-job/examples/simple/jobs"
	"github.com/gin-job/gin-job/job"
	"github.com/gin-job/gin-job/router"
)

func main() {
	// register job
	jobList := []job.Job{
		&jobs.ExampleJob{},
	}

	// init router
	r := router.NewGinJobRouter(nil)
	r.SetJobList(jobList)
	r.Start()
}
