package job

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]Job)
	mu       sync.RWMutex
)

// Register 注册一个任务实现
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

// Get 获取已注册的任务实现
func Get(name string) (Job, error) {
	mu.RLock()
	defer mu.RUnlock()

	j, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("任务 %s 不存在", name)
	}

	return j, nil
}

// List 列出所有已注册的任务名称和描述
func List() map[string]string {
	mu.RLock()
	defer mu.RUnlock()

	result := make(map[string]string)
	for name, j := range registry {
		result[name] = j.Description()
	}

	return result
}

// Unregister 取消注册任务
func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(registry, name)
}
