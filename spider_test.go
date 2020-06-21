package gospider

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zhshch2002/goreq"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSpider(t *testing.T) {
	a := 0
	s := NewSpider(func(s *Spider) {
		a += 1
	})
	s.OnTask(func(ctx *Context, t *Task) *Task {
		fmt.Println("OnTask")
		a += 1
		return t
	})
	s.OnResp(func(ctx *Context) {
		fmt.Println("OnResp")
		a += 1
	})
	s.OnItem(func(ctx *Context, i interface{}) interface{} {
		fmt.Println("OnItem", i)
		a += 1
		return i
	})
	s.OnRecover(func(ctx *Context, err error) {
		a += 1
		fmt.Println("OnRecover", err)
	})
	s.OnReqError(func(ctx *Context, err error) {
		a += 1
		fmt.Println("OnReqError", err)
	})
	s.OnRespError(func(ctx *Context, err error) {
		a += 1
		fmt.Println("OnRespError", err)
	})

	s.SeedTask(
		goreq.Get("https://httpbin.org/get"),
		func(ctx *Context) {
			ctx.AddItem(ctx.Resp.Text)
			panic("test panic")
		},
	)

	r := goreq.Get("https://httpbin.org/get")
	r.Err = errors.New("test error")
	s.SeedTask(r)

	s.SeedTask(goreq.Get("htps://httpbin.org/get"))

	s.Wait()
	assert.Equal(t, 9, a)
}

func TestContext_Abort(t *testing.T) {
	c := make(chan struct{})
	s := NewSpider()
	s.SeedTask(
		goreq.Get("https://httpbin.org/get"),
		func(ctx *Context) {
			ctx.AddItem(ctx.Resp.Text)
			ctx.Abort()
			c <- struct{}{}
		},
		func(ctx *Context) {
			t.Error("abort fail")
		},
	)
	_ = <-c
	s.Wait()
}

func TestSpiderManyTask(t *testing.T) {
	s := NewSpider()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	i := 0
	a := 30
	for a > 0 {
		s.SeedTask(
			goreq.Get(ts.URL),
			func(ctx *Context) {
				i += 1
			},
		)
		a -= 1
	}
	s.Wait()
	assert.Equal(t, 30, i)
}