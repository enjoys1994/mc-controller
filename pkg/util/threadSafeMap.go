package util

import "sync"

type ThreadSafeMap struct {
	values map[interface{}]interface{}
	mu     sync.RWMutex
}

func (t *ThreadSafeMap) Load(i interface{}) (interface{}, bool) {
	t.mu.RLock()
	if t.values == nil {
		t.values = map[interface{}]interface{}{}
	}
	v, ok := t.values[i]
	t.mu.RUnlock()
	return v, ok
}

func (t *ThreadSafeMap) Store(key, value interface{}) {
	t.mu.Lock()
	if t.values == nil {
		t.values = map[interface{}]interface{}{}
	}
	t.values[key] = value
	t.mu.Unlock()
}

func (t *ThreadSafeMap) Delete(i interface{}) {
	delete(t.values, i)
}

func (t *ThreadSafeMap) Size() int {
	return len(t.values)
}

func (t *ThreadSafeMap) Range(f func(key interface{}, value interface{}) (shouldContinue bool)) {
	for k, v := range t.values {
		shouldContinue := f(k, v)
		if !shouldContinue {
			return
		}
	}
}
