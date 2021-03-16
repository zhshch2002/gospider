package gospider

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zhshch2002/goreq"
	"net/http"
	"testing"
)

func TestWithDeduplicate(t *testing.T) {
	s := NewSpider(WithDeduplicate())
	got := false
	s.AddRootTask(goreq.Get("https://httpbin.org/get").AddParam("a", "a").AddCookie(&http.Cookie{
		Name:  "b",
		Value: "b",
	}).SetRawBody([]byte("c=c")).AddHeader("d", "d"), func(ctx *Task) {
		got = true
	})
	s.AddRootTask(goreq.Get("https://httpbin.org/get").AddParam("a", "a").AddCookie(&http.Cookie{
		Name:  "b",
		Value: "b",
	}).SetRawBody([]byte("c=c")).AddHeader("d", "d"), func(ctx *Task) {
		t.Error("Deduplicate error")
	})
	s.Wait()
	assert.True(t, got)
}

func TestWithRobotsTxt(t *testing.T) {
	s := NewSpider(WithRobotsTxt("gospider"))
	s.AddRootTask(goreq.Get("https://github.com/gist/"), func(ctx *Task) {
		t.Error("RobotsTxt error")
	})
	got := false
	s.AddRootTask( // unable to access according to https://github.com/robots.txt
		goreq.Get("https://github.com/"),
		func(ctx *Task) {
			got = true
		},
	)
	s.Wait()
	assert.True(t, got)
}

func TestWithDepthLimit(t *testing.T) {
	s := NewSpider(WithDepthLimit(2))
	s.AddRootTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
		ctx.Println("Depth", ctx.Req.Context().Value("depth")) // 1
		ctx.AddTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
			ctx.Println("Depth", ctx.Req.Context().Value("depth")) // 2
			ctx.AddTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
				ctx.Println("Depth", ctx.Req.Context().Value("depth")) // 3
				t.Error("Limiter error")
			})
		})
	})
	s.Wait()
}

func TestWithMaxReqLimit(t *testing.T) {
	s := NewSpider(WithMaxReqLimit(2))
	count := 0
	s.AddRootTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
		count += 1
	})
	s.AddRootTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
		count += 1
	})
	s.AddRootTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
		count += 1
	})
	s.Wait()
	assert.Equal(t, 2, count)
}

func TestWithErrorLog(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	s := NewSpider(WithErrorLog(buf))
	s.AddRootTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
		ctx.AddItem(errors.New("test item error"))
		panic("test panic error")
	})
	s.Wait()
	fmt.Println(buf.String())
	assert.True(t, buf.Len() > 0)
}

func TestWithCsvItemSaver(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	s := NewSpider(WithCsvItemSaver(buf))
	s.AddRootTask(goreq.Get("https://httpbin.org/get"), func(ctx *Task) {
		ctx.AddItem(CsvItem{ctx.Request.URL.String(), ctx.Status})
	})
	s.Wait()
	fmt.Println(buf.String())
	assert.True(t, buf.Len() > 0)
}
