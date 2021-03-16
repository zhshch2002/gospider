package gospider

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zhshch2002/goreq"
	"strings"
)

type Task struct {
	*goreq.Response
	s        *Spider
	Handlers []Handler
	Meta     map[string]interface{}
	abort    bool
}

// Abort this context to break the handler chain and stop handling
func (c *Task) Abort() {
	c.abort = true
}

// IsAborted return was the context dropped
func (c *Task) IsAborted() bool {
	return c.abort
}

// AddTask start a new goroutine to execute network requests and handler chain
func (c *Task) AddTask(req *goreq.Request, h ...Handler) {
	if !req.URL.IsAbs() {
		req.URL = c.Req.URL.ResolveReference(req.URL)
	}
	t := c.s.handleOnTask(c, NewTask(req, c.s, MetaCopy(c.Meta), h...))
	if t == nil {
		return
	}
	c.s.addTask(t)
}

// AddItem start a new goroutine to execute the OnItem handler chain
func (c *Task) AddItem(i interface{}) {
	c.s.addItem(&Item{
		Task: c,
		Data: i,
	})
}

// IsDownloaded return true if the request is downloaded
func (c *Task) IsDownloaded() bool {
	return c.Response != nil
}

func (c *Task) Println(v ...interface{}) {
	Logger.Info().Str("spider", c.s.Name).Str("task", c.String()).Msg(strings.TrimRight(fmt.Sprintln(v...), "\n"))
}

func (c *Task) Debug(v ...interface{}) {
	Logger.Debug().Str("spider", c.s.Name).Str("task", c.String()).Msg(strings.TrimRight(fmt.Sprintln(v...), "\n"))
}

func (c *Task) Error(err error) {
	Logger.Err(err).Stack().Str("spider", c.s.Name).Str("task", c.String()).Send()
}

func (c *Task) Logger() zerolog.Logger {
	return Logger
}

func (c *Task) String() string {
	if c.Req == nil {
		return "[empty task]"
	} else if c.Response == nil {
		return fmt.Sprint("[not downloaded task] ", c.Req.URL.String())
	} else if c.Response.Response == nil || c.Err != nil {
		return fmt.Sprint("[err task] ", c.Req.URL.String())
	} else {
		return fmt.Sprint("["+c.Status+"] ", c.Req.URL)
	}
}
