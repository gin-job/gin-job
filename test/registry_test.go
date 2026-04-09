package test

import (
	"context"
	"testing"

	"github.com/gin-job/gin-job/docs/assets/job"
)

// Test Job implementation
type testJob struct {
	name        string
	description string
}

func (t *testJob) Name() string {
	return t.name
}

func (t *testJob) Run(ctx context.Context) error {
	return nil
}

func (t *testJob) Description() string {
	return t.description
}

func TestRegister(t *testing.T) {
	// Test registering a new job
	job1 := &testJob{name: "test1", description: "Test Job 1"}
	err := job.Register(job1)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// Test registering a job with duplicate name
	err = job.Register(job1)
	if err == nil {
		t.Errorf("Register should fail for duplicate job")
	}

	// Test registering a job with empty name
	job2 := &testJob{name: "", description: "Test Job 2"}
	err = job.Register(job2)
	if err == nil {
		t.Errorf("Register should fail for empty job name")
	}

	// Cleanup
	job.Unregister("test1")
}

func TestGet(t *testing.T) {
	// Register test job
	job1 := &testJob{name: "test1", description: "Test Job 1"}
	err := job.Register(job1)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// Test getting registered job
	j, err := job.Get("test1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if j.Name() != "test1" {
		t.Errorf("Get returned wrong job name: %s", j.Name())
	}

	// Test getting non-existent job
	_, err = job.Get("non-existent")
	if err == nil {
		t.Errorf("Get should fail for non-existent job")
	}

	// 清理
	job.Unregister("test1")
}

func TestList(t *testing.T) {
	// Register test jobs
	job1 := &testJob{name: "test1", description: "Test Job 1"}
	job2 := &testJob{name: "test2", description: "Test Job 2"}
	err := job.Register(job1)
	if err != nil {
		t.Errorf("Register job1 failed: %v", err)
	}
	err = job.Register(job2)
	if err != nil {
		t.Errorf("Register job2 failed: %v", err)
	}

	// Test listing all jobs
	jobs := job.List()
	if len(jobs) != 2 {
		t.Errorf("List returned wrong number of jobs: %d", len(jobs))
	}
	if jobs["test1"] != "Test Job 1" {
		t.Errorf("List returned wrong description for test1: %s", jobs["test1"])
	}
	if jobs["test2"] != "Test Job 2" {
		t.Errorf("List returned wrong description for test2: %s", jobs["test2"])
	}

	// 清理
	job.Unregister("test1")
	job.Unregister("test2")
}

func TestUnregister(t *testing.T) {
	// Register test job
	job1 := &testJob{name: "test1", description: "Test Job 1"}
	err := job.Register(job1)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// Test unregistering job
	job.Unregister("test1")

	// Test getting unregistered job
	_, err = job.Get("test1")
	if err == nil {
		t.Errorf("Get should fail for unregistered job")
	}

	// Test unregistering non-existent job (should not error)
	job.Unregister("non-existent")
}
