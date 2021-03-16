package gospider

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/slyrz/robots"
	"github.com/zhshch2002/goreq"
	"io"
	"strings"
	"sync"
	"sync/atomic"
)

func WithDeduplicate() Extension {
	return func(s *Spider) {
		CrawledHash := map[string]struct{}{}
		lock := sync.Mutex{}
		s.OnTask(func(o, t *Task) *Task {
			has := goreq.GetRequestHash(t.Req)
			lock.Lock()
			defer lock.Unlock()
			if _, ok := CrawledHash[has]; ok {
				return nil
			}
			CrawledHash[has] = struct{}{}
			return t
		})
	}

}

func WithRobotsTxt(ua string) Extension {
	return func(s *Spider) {
		if ua == "" {
			ua = s.Name
		}
		rs := map[string]*robots.Robots{}
		s.OnTask(func(o, t *Task) *Task {
			var r *robots.Robots
			if a, ok := rs[t.Req.URL.Host]; ok {
				r = a
			} else {
				if u, err := t.Req.URL.Parse("/robots.txt"); err == nil {
					if resp, err := goreq.Get(u.String()).SetUA(ua).Do().Resp(); err == nil && resp.StatusCode == 200 {
						r = robots.New(strings.NewReader(resp.Text), ua)
						rs[t.Req.URL.Host] = r
					}
				}
			}
			if r != nil {
				if !r.Allow(t.Req.URL.Path) {
					return nil
				}
			}
			return t
		})
	}
}

func WithDepthLimit(max int) Extension {
	return func(s *Spider) {
		s.OnTask(func(o, t *Task) *Task {
			if o.Response == nil || o.Req == nil || o.Req.Context().Value("depth") == nil {
				t.Req.Request = t.Req.WithContext(context.WithValue(t.Req.Context(), "depth", 1))
				return t
			} else {
				depth := o.Req.Context().Value("depth").(int)
				if depth < max {
					t.Req.Request = t.Req.WithContext(context.WithValue(t.Req.Context(), "depth", depth+1))
					return t
				} else {
					return nil
				}
			}
		})
	}
}

func WithMaxReqLimit(max int64) Extension {
	return func(s *Spider) {
		count := int64(0)
		s.OnTask(func(o, t *Task) *Task {
			if count < max {
				atomic.AddInt64(&count, 1)
				return t
			}
			return nil
		})
	}
}

func WithErrorLog(f io.Writer) Extension {
	return func(s *Spider) {
		l := zerolog.New(f).With().Timestamp().Stack().Logger()
		send := func(task *Task, err error, t, stack string) {
			event := l.Err(err).
				Stack().
				Str("spider", s.Name).
				Str("type", "item").
				Str("task", fmt.Sprint(task)).
				Str("url", task.Req.URL.String()).
				AnErr("req err", task.Req.Err).
				AnErr("resp err", task.Response.Err)
			if task.Response != nil {
				event.Int("resp code", task.Response.StatusCode)
				if task.Response.Text != "" {
					event.Str("text", task.Response.Text)
				}
			}
			event.Str("stack", SprintStack()).Send()
		}

		s.OnItem(func(ctx *Task, i interface{}) interface{} {
			if err, ok := i.(error); ok {
				send(ctx, err, "item", SprintStack())
			}
			return i
		})
		s.OnRecover(func(ctx *Task, err error) {
			send(ctx, err, "OnRecover", SprintStack())
		})
		s.OnReqError(func(ctx *Task, err error) {
			send(ctx, err, "OnReqError", SprintStack())
		})
		s.OnRespError(func(ctx *Task, err error) {
			send(ctx, err, "OnRespError", SprintStack())
		})
	}
}

type CsvItem []string

func WithCsvItemSaver(f io.Writer) Extension {
	lock := sync.Mutex{}
	w := csv.NewWriter(f)
	return func(s *Spider) {
		s.OnItem(func(ctx *Task, i interface{}) interface{} {
			if data, ok := i.(CsvItem); ok {
				lock.Lock()
				defer lock.Unlock()
				err := w.Write(data)
				if err != nil {
					Logger.Err(err).Msg("WithCsvItemSaver Error")
				}
				w.Flush()
			}
			return i
		})
	}
}
