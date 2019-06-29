# cod-router-concurrent-limiter

[![Build Status](https://img.shields.io/travis/vicanso/cod-router-concurrent-limiter.svg?label=linux+build)](https://travis-ci.org/vicanso/cod-router-concurrent-limiter)


Router concurrent limiter for cod, it support custom max concurrency for each router.

- `NewLimiter` create a limiter for router concurrent limit.

```go
package main

import (
	"time"

	"github.com/vicanso/cod"

	responder "github.com/vicanso/cod-responder"
	routerLimiter "github.com/vicanso/cod-router-concurrent-limiter"
)

func main() {
	d := cod.New()

	d.Use(routerLimiter.New(routerLimiter.Config{
		Limiter: routerLimiter.NewLimiter(map[string]uint32{
			"/users/me": 2,
		}),
	}))
	d.Use(responder.NewDefault())

	d.GET("/users/me", func(c *cod.Context) (err error) {
		time.Sleep(time.Second)
		c.Body = map[string]string{
			"account": "tree",
			"name":    "tree.xie",
		}
		return nil
	})
	d.ListenAndServe(":3000")
}

```