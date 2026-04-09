package job

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]Job)
	mu       sync.RWMutex
)

// Register Job
func Register(j Job) error {
	mu.Lock()
	defer mu.Unlock()

	name := j.Name()
	if name == "" {
		return fmt.Errorf("任务名不能为空")
	}

	if _, exists := registry[name]; exists {
		return fmt.Errorf("任务 %s 已经创建过", name)
	}

	registry[name] = j
	return nil
}

// Get Job
func Get(name string) (Job, error) {
	mu.RLock()
	defer mu.RUnlock()

	j, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("任务 %s 不存在", name)
	}

	return j, nil
}

// List Jobs
func List() map[string]string {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]string)
	for name, j := range registry {
		result[name] = j.Description()
	}

	return result
}

// Unregister Job
func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(registry, name)
}
