package main

import (
	"bytes"
	"time"

	"github.com/vicanso/elton"

	routerLimiter "github.com/vicanso/elton-router-concurrent-limiter"
)

func main() {
	e := elton.New()

	e.Use(routerLimiter.New(routerLimiter.Config{
		Limiter: routerLimiter.NewLocalLimiter(map[string]uint32{
			"GET /users/me": 2,
		}),
	}))

	e.GET("/users/me", func(c *elton.Context) (err error) {
		time.Sleep(time.Second)
		c.BodyBuffer = bytes.NewBufferString(`{
			"account": "tree",
			"name": "tree.xie"
		}`)
		return nil
	})
	err := e.ListenAndServe(":3000")
	if err != nil {
		panic(err)
	}
}
