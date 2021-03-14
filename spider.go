package gospider

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/zhshch2002/goreq"
	"sync"
)

var (
	UnknownExt = errors.New("unknown ext")
)

type Handler func(ctx *Task)
type Extension func(s *Spider)

type Item struct {
	Ctx  *Task
	Data interface{}
}

func NewTask(req *goreq.Request, s *Spider, meta map[string]interface{}, a ...Handler) (t *Task) {
	t = &Task{
		Response: &goreq.Response{
			Req: req,
		},
		s:        s,
		Handlers: a,
		Meta:     meta,
		abort:    false,
	}
	return
}

type Spider struct {
	Name       string
	Logging    bool
	SyncOnItem bool //TODO

	Client *goreq.Client
	wg     sync.WaitGroup

	onTaskHandlers      []func(t *Task) *Task
	onRespHandlers      []Handler
	onItemHandlers      []func(ctx *Task, i interface{}) interface{}
	onRecoverHandlers   []func(ctx *Task, err error)
	onReqErrorHandlers  []func(ctx *Task, err error)
	onRespErrorHandlers []func(ctx *Task, err error)
}

func NewSpider(e ...interface{}) *Spider {
	s := &Spider{
		Name:       "gospider",
		Logging:    true,
		SyncOnItem: false,
		Client:     goreq.NewClient(),
		wg:         sync.WaitGroup{},
	}
	s.Use(e...)
	return s
}

func (s *Spider) Use(exts ...interface{}) {
	for _, fn := range exts {
		switch fn.(type) {
		case func(s *Spider):
			fn.(func(s *Spider))(s)
			break
		case Extension:
			fn.(Extension)(s)
			break
		case goreq.Middleware, func(*goreq.Client, goreq.Handler) goreq.Handler:
			s.Client.Use(fn.(goreq.Middleware))
			break
		default:
			panic(UnknownExt)
		}
	}
}

func (s *Spider) Forever() {
	select {}
}

func (s *Spider) Wait() {
	s.wg.Wait()
}

func (s *Spider) handleTask(t *Task) {
	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				if s.Logging {
					Logger.Error().
						Stack().
						Err(errors.WithStack(e)).
						Str("spider", s.Name).
						Str("task", fmt.Sprint(t)).
						Msg("handler recover from panic")
				}
				s.handleOnError(t, e)
			} else {
				if s.Logging {
					Logger.Error().
						Stack().
						Err(errors.WithStack(fmt.Errorf("%v", err))).
						Str("spider", s.Name).
						Str("task", fmt.Sprint(t)).
						Msg("handler recover from panic")
				}
				s.handleOnError(t, fmt.Errorf("%v", err))
			}
		}
	}()

	if t.Req.Err != nil {
		if s.Logging {
			Logger.Error().
				Stack().
				Err(t.Req.Err).
				Str("spider", s.Name).
				Str("task", fmt.Sprint(t)).
				Msg("req error")
		}
		s.handleOnReqError(t, t.Req.Err)
		return
	}

	t.Response = s.Client.Do(t.Req)
	if t.Err != nil {
		if s.Logging {
			Logger.Error().
				Stack().
				Err(t.Err).
				Str("spider", s.Name).
				Str("task", fmt.Sprint(t)).
				Msg("resp error")
		}
		s.handleOnRespError(t, t.Err)
		return
	}

	if s.Logging {
		Logger.Debug().
			Str("spider", s.Name).
			Str("context", fmt.Sprint(t)).
			Msg("finish")
	}

	s.handleOnResp(t)
	if t.IsAborted() {
		return
	}
	for _, fn := range t.Handlers {
		fn(t)
		if t.IsAborted() {
			return
		}
	}
}

func (s *Spider) StartFrom(req *goreq.Request, h ...Handler) {
	s.addTask(NewTask(req, s, map[string]interface{}{}, h...))
}

func (s *Spider) addTask(t *Task) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleTask(t)
	}()
}

func (s *Spider) addItem(i *Item) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleOnItem(i)
	}()
}

/*************************************************************************************/
func (s *Spider) OnTask(fn func(t *Task) *Task) {
	s.onTaskHandlers = append(s.onTaskHandlers, fn)
}
func (s *Spider) handleOnTask(t *Task) *Task {
	for _, fn := range s.onTaskHandlers {
		t = fn(t)
		if t == nil || t.IsAborted() {
			return nil
		}
	}
	return t
}

/*************************************************************************************/
func (s *Spider) OnResp(fn Handler) {
	s.onRespHandlers = append(s.onRespHandlers, fn)
}
func (s *Spider) OnHTML(selector string, fn func(t *Task, sel *goquery.Selection)) {
	s.OnResp(func(t *Task) {
		if t.IsHTML() {
			if h, err := t.HTML(); err == nil {
				h.Find(selector).Each(func(i int, selection *goquery.Selection) {
					fn(t, selection)
				})
			}
		}
	})
}
func (s *Spider) OnJSON(q string, fn func(t *Task, j gjson.Result)) {
	s.onRespHandlers = append(s.onRespHandlers, func(t *Task) {
		if t.IsJSON() {
			if j, err := t.JSON(); err == nil {
				if res := j.Get(q); res.Exists() {
					fn(t, res)
				}
			}
		}
	})
}
func (s *Spider) handleOnResp(t *Task) {
	for _, fn := range s.onRespHandlers {
		if t.IsAborted() {
			return
		}
		fn(t)
	}
}

/*************************************************************************************/
func (s *Spider) OnItem(fn func(t *Task, i interface{}) interface{}) {
	s.onItemHandlers = append(s.onItemHandlers, fn)
}
func (s *Spider) handleOnItem(i *Item) {
	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				if s.Logging {
					Logger.Error().
						Stack().
						Err(errors.WithStack(e)).
						Str("spider", s.Name).
						Str("context", fmt.Sprint(i.Ctx)).
						Msg("handler recover from panic")
				}
				s.handleOnError(i.Ctx, e)
			} else {
				if s.Logging {
					Logger.Error().
						Stack().
						Err(errors.WithStack(fmt.Errorf("%v", err))).
						Str("spider", s.Name).
						Str("context", fmt.Sprint(i.Ctx)).
						Msg("handler recover from panic")
				}
				s.handleOnError(i.Ctx, fmt.Errorf("%v", err))
			}
		}
	}()
	for _, fn := range s.onItemHandlers {
		i.Data = fn(i.Ctx, i.Data)
		if i.Data == nil {
			return
		}
	}
}

/*************************************************************************************/
func (s *Spider) OnRecover(fn func(t *Task, err error)) {
	s.onRecoverHandlers = append(s.onRecoverHandlers, fn)
}
func (s *Spider) handleOnError(t *Task, err error) {
	for _, fn := range s.onRecoverHandlers {
		fn(t, err)
	}
}

func (s *Spider) OnRespError(fn func(t *Task, err error)) {
	s.onRespErrorHandlers = append(s.onRespErrorHandlers, fn)
}
func (s *Spider) handleOnRespError(t *Task, err error) {
	for _, fn := range s.onRespErrorHandlers {
		fn(t, err)
	}
}

func (s *Spider) OnReqError(fn func(t *Task, err error)) {
	s.onReqErrorHandlers = append(s.onReqErrorHandlers, fn)
}
func (s *Spider) handleOnReqError(t *Task, err error) {
	for _, fn := range s.onReqErrorHandlers {
		fn(t, err)
	}
}
