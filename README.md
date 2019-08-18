# elton-router-concurrent-limiter

[![Build Status](https://img.shields.io/travis/vicanso/elton-router-concurrent-limiter.svg?label=linux+build)](https://travis-ci.org/vicanso/elton-router-concurrent-limiter)


Router concurrent limiter for elton, it support custom max concurrency for each router.

- `NewLocalLimiter` create a limiter for router concurrent limit.

```go
package main

import (
	"time"

	"github.com/vicanso/elton"

	responder "github.com/vicanso/elton-responder"
	routerLimiter "github.com/vicanso/elton-router-concurrent-limiter"
)

func main() {
	d := elton.New()

	d.Use(routerLimiter.New(routerLimiter.Config{
		Limiter: routerLimiter.NewLocalLimiter(map[string]uint32{
			"GET /users/me": 2,
		}),
	}))
	d.Use(responder.NewDefault())

	d.GET("/users/me", func(c *elton.Context) (err error) {
		time.Sleep(time.Second)
		c.Body = map[string]string{
			"account": "tree",
			"name":    "tree.xie",
		}
		return nil
	})
	d.ListenAndServe(":7001")
}
```