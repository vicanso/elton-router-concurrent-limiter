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
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vicanso/cod"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	assert := assert.New(t)
	limiter := NewLimiter(map[string]uint32{
		"/users/login": 10,
		"/books/:id":   100,
	})

	cur, max := limiter.IncConcurrency("/not-macth-route")
	assert.Equal(uint32(0), max)
	assert.Equal(uint32(0), cur)

	cur, max = limiter.IncConcurrency("/users/login")
	assert.Equal(uint32(10), max)
	assert.Equal(uint32(1), cur)

	limiter.DecConcurrency("/not-macth-route")
	assert.Equal(uint32(0), limiter.GetConcurrency("/not-macth-route"))

	limiter.DecConcurrency("/users/login")
	assert.Equal(uint32(0), limiter.GetConcurrency("/users/login"))
}

func TestNoLimiterPanic(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		r := recover()
		assert.NotNil(r)
		assert.Equal(r.(error), errRequireLimiter)
	}()

	New(Config{})
}

func TestRouterConcurrentLimiter(t *testing.T) {
	limiter := NewLimiter(map[string]uint32{
		"/users/login": 1,
		"/books/:id":   100,
	})
	fn := New(Config{
		Limiter: limiter,
	})
	t.Run("skip", func(t *testing.T) {
		assert := assert.New(t)
		c := cod.NewContext(nil, nil)
		c.Committed = true
		done := false
		c.Next = func() error {
			done = true
			return nil
		}
		err := fn(c)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("below limit", func(t *testing.T) {
		assert := assert.New(t)
		c := cod.NewContext(nil, nil)
		c.Route = "/books/:id"
		var count int32
		max := 10
		c.Next = func() error {
			atomic.AddInt32(&count, 1)
			return nil
		}

		for index := 0; index < max; index++ {
			err := fn(c)
			assert.Nil(err)
		}
		assert.Equal(int32(max), count)
	})

	t.Run("higher than limit", func(t *testing.T) {
		assert := assert.New(t)
		c := cod.NewContext(nil, nil)
		c.Route = "/users/login"
		c.Next = func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		done := make(chan bool)
		go func() {
			time.Sleep(2 * time.Millisecond)
			err := fn(c)
			assert.NotNil(err)
			assert.Equal("category=cod-router-concurrent-limiter, message=too many requset, current:2, max:1", err.Error())
			done <- true
		}()
		err := fn(c)
		assert.Nil(err)
		<-done
	})
}

// https://stackoverflow.com/questions/50120427/fail-unit-tests-if-coverage-is-below-certain-percentage
func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	rc := m.Run()

	// rc 0 means we've passed,
	// and CoverMode will be non empty if run with -cover
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 0.9 {
			fmt.Println("Tests passed but coverage failed at", c)
			rc = -1
		}
	}
	os.Exit(rc)
}
