package gospider

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/zhshch2002/goreq"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSpider_OnResp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	s := NewSpider()

	a := 0

	s.OnResp(func(task *Task) {
		task.Println("OnResp", task.Text)
		a += 1
	})

	s.AddRootTask(
		goreq.Get(ts.URL),
		func(task *Task) {
			task.Println("Handler")
			a += 1
		},
	)
	s.Wait()
	assert.Equal(t, 2, a)
}

func TestSpider_OnTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	s := NewSpider()

	a := 0

	s.OnTask(func(o, task *Task) *Task {
		task.Println("OnTask", task.Req.URL.String())
		a += 1
		if task.Req.URL.String() != ts.URL {
			task.Println("drop task")
			return nil
		}
		task.Meta["hello"] = "gospider"
		return task
	})

	s.AddRootTask(
		goreq.Get(ts.URL),
		func(task *Task) {
			task.Println("Handler", task.Meta["hello"])
			assert.Equal(t, "gospider", task.Meta["hello"])
			a += 1
		},
	)

	s.AddRootTask(
		goreq.Get("https://httpbin.org/get"),
		func(task *Task) {
			task.Println("Handler 2")
			a += 1
		},
	)
	s.Wait()
	assert.Equal(t, 3, a)
}

func TestSpider_OnRecover(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	s := NewSpider()

	a := 0

	s.OnReqError(func(task *Task, err error) {
		task.Println("OnReqError", err)
		a += 1
	})

	s.OnRespError(func(task *Task, err error) {
		task.Println("OnRespError", err)
		a += 1
	})

	s.OnRecover(func(task *Task, err error) {
		task.Println("OnRecover", err)
		a += 1
	})

	req := goreq.Get(ts.URL)
	req.Err = errors.New("test req error")
	s.AddRootTask(
		req,
		func(task *Task) {
			a += 1
			t.Fatal("error")
		},
	)
	s.AddRootTask(
		goreq.Get("htta://localhost"),
		func(task *Task) {
			a += 1
			t.Fatal("error")
		},
	)
	s.AddRootTask(
		goreq.Get(ts.URL),
		func(task *Task) {
			a += 1
			panic(errors.New("test panic error"))
			t.Fatal("error")
		},
	)

	s.Wait()
	assert.Equal(t, 4, a)
}

func TestSpider_OnItem(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	s := NewSpider()

	a := 0

	s.OnItem(func(task *Task, i interface{}) interface{} {
		a += 1
		assert.Equal(t, "Hello", i)
		return i
	})

	s.AddRootTask(
		goreq.Get(ts.URL),
		func(task *Task) {
			a += 1
			task.AddItem(task.Text)
		},
	)

	s.Wait()
	assert.Equal(t, 2, a)
}

func TestTask_Abort(t *testing.T) {
	c := make(chan struct{})
	s := NewSpider()
	s.AddRootTask(
		goreq.Get("https://httpbin.org/get"),
		func(task *Task) {
			task.AddItem(task.Text)
			task.Abort()
			c <- struct{}{}
		},
		func(task *Task) {
			t.Fatal("abort fail")
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
		s.AddRootTask(
			goreq.Get(ts.URL),
			func(t *Task) {
				i += 1
			},
		)
		a -= 1
	}
	s.Wait()
	assert.Equal(t, 30, i)
}

func BenchmarkSpider(b *testing.B) {
	s := NewSpider()
	s.Logging = false
	req := goreq.Get("http://127.0.0.1:8080/")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.AddRootTask(req)
	}
	s.Wait()
}
func BenchmarkGoreq(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello")
	}))
	defer ts.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		goreq.Get(ts.URL).Do()
	}
}
