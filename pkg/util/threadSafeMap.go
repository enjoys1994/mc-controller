package util

import "sync"

var _ mapInterface = (*ThreadSafeMap)(nil)

type ThreadSafeMap struct {
	values map[interface{}]interface{}
	mu     sync.RWMutex
}

func (t ThreadSafeMap) Load(i interface{}) (interface{}, bool) {
	t.mu.RLock()
	if t.values == nil {
		t.values = map[interface{}]interface{}{}
	}
	v, ok := t.values[i]
	t.mu.RUnlock()
	return v, ok
}

func (t ThreadSafeMap) Store(key, value interface{}) {
	t.mu.Lock()
	if t.values == nil {
		t.values = map[interface{}]interface{}{}
	}
	t.values[key] = value
	t.mu.Unlock()
}

func (t ThreadSafeMap) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {

	t.mu.Lock()
	if t.values == nil {
		t.values = map[interface{}]interface{}{}
	}
	v, ok := t.values[key]
	if ok {
		loaded = true
		actual = v
	} else {
		t.values[key] = value
		loaded = false
		actual = value
	}
	t.mu.Unlock()
	return actual, loaded
}

func (t ThreadSafeMap) Delete(i interface{}) {
	delete(t.values, i)
}

func (t ThreadSafeMap) Range(f func(key interface{}, value interface{}) (shouldContinue bool)) {
	for k, v := range t.values {
		shouldContinue := f(k, v)
		if !shouldContinue {
			return
		}
	}
}

type mapInterface interface {
	Load(interface{}) (interface{}, bool)
	Store(key, value interface{})
	LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
	Delete(interface{})
	Range(func(key, value interface{}) (shouldContinue bool))
}
