// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routerconcurrentlimiter

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/vicanso/hes"

	"github.com/vicanso/cod"
)

const (
	// ErrCategory router concurrent limiter error category
	ErrCategory = "cod-router-concurrent-limiter"
)

var (
	errRequireLimiter = errors.New("require limiter")
)

type (
	// Config router concurrent limiter config
	Config struct {
		Skipper cod.Skipper
		Limiter Limiter
	}
	concurrency struct {
		max     uint32
		current uint32
	}
	// Limiter limiter interface
	Limiter interface {
		IncConcurrency(route string) (current uint32, max uint32)
		DecConcurrency(route string)
		GetConcurrency(route string) (current uint32)
	}
	// LocalLimiter local limiter
	LocalLimiter struct {
		m map[string]*concurrency
	}
)

// NewLocalLimiter create a new limiter
func NewLocalLimiter(data map[string]uint32) *LocalLimiter {
	m := make(map[string]*concurrency)
	for route, max := range data {
		m[route] = &concurrency{
			max:     max,
			current: 0,
		}
	}
	return &LocalLimiter{
		m: m,
	}
}

// IncConcurrency concurrency inc
func (l *LocalLimiter) IncConcurrency(route string) (current, max uint32) {
	concur, ok := l.m[route]
	if !ok {
		return 0, 0
	}
	v := atomic.AddUint32(&concur.current, 1)
	return v, concur.max
}

// DecConcurrency concurrency dec
func (l *LocalLimiter) DecConcurrency(route string) {
	concur, ok := l.m[route]
	if !ok {
		return
	}
	atomic.AddUint32(&concur.current, ^uint32(0))
}

// GetConcurrency get concurrency
func (l *LocalLimiter) GetConcurrency(route string) uint32 {
	concur, ok := l.m[route]
	if !ok {
		return 0
	}
	return atomic.LoadUint32(&concur.current)
}

func createError(current, max uint32) error {
	he := hes.New(fmt.Sprintf("too many requset, current:%d, max:%d", current, max))
	he.Category = ErrCategory
	he.StatusCode = http.StatusTooManyRequests
	return he
}

// New create a concurrent limiter middleware
func New(config Config) cod.Handler {
	skipper := config.Skipper
	if skipper == nil {
		skipper = cod.DefaultSkipper
	}
	if config.Limiter == nil {
		panic(errRequireLimiter)
	}
	limiter := config.Limiter
	return func(c *cod.Context) (err error) {
		if skipper(c) {
			return c.Next()
		}
		route := c.Route
		current, max := limiter.IncConcurrency(route)
		defer limiter.DecConcurrency(route)
		if current > max {
			err = createError(current, max)
			return
		}
		return c.Next()
	}
}
