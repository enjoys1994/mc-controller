/*
Copyright 2018 The Multicluster-Controller Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"context"
	"fmt"
	"sync"
)

// ControllerSet is a set of Controllers.

// Manager manages controllers. It starts their caches, waits for those to sync, then starts the controllers.
type Manager struct {
	controllers []Controller
}

// New creates a Manager.
func New() *Manager {
	return &Manager{controllers: []Controller{}}
}

// Cache is the interface used by Manager to start and wait for caches to sync.
type Cache interface {
	Start(ctx context.Context) error
	WaitForCacheSync(ctx context.Context) bool
}

// Controller is the interface used by Manager to start the controllers and get their caches (beforehand).
type Controller interface {
	Start(ctx context.Context) error
	GetCaches() []Cache
}

// AddController adds a controller to the Manager.
func (m *Manager) AddController(c Controller) {
	m.controllers = append(m.controllers, c)
}

// Start gets all the unique caches of the controllers it manages, starts them,
// then starts the controllers as soon as their respective caches are synced.
// Start blocks until an error or stop is received.
func (m *Manager) Start(ctx context.Context) error {
	errCh := make(chan error)

	wgs := make(map[Controller]*sync.WaitGroup)
	caches := make(map[Cache][]Controller)

	for i := range m.controllers {
		controller := m.controllers[i]
		wgs[controller] = &sync.WaitGroup{}
		for i := range controller.GetCaches() {
			ca := controller.GetCaches()[i]
			wgs[controller].Add(1)
			cos, ok := caches[ca]
			if !ok {
				cos = []Controller{}
			}
			cos = append(cos, controller)
			caches[ca] = cos
		}
	}

	for ca, cos := range caches {
		go func(ca Cache) {
			if err := ca.Start(ctx); err != nil {
				errCh <- err
			}
		}(ca)
		go func(ca Cache, controllers []Controller) {
			if ok := ca.WaitForCacheSync(ctx); !ok {
				errCh <- fmt.Errorf("failed to wait for caches to sync")
			}
			for i := range controllers {
				wgs[controllers[i]].Done()
			}
		}(ca, cos)
	}

	for i := range m.controllers {
		co := m.controllers[i]
		go func(co Controller) {
			wgs[co].Wait()
			if err := co.Start(ctx); err != nil {
				errCh <- err
			}
		}(co)
	}

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}
