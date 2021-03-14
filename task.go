package gospider

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/zhshch2002/goreq"
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
	t := c.s.handleOnTask(NewTask(req, c.s, c.Meta, h...))
	if t == nil {
		return
	}
	c.s.addTask(t)
}

// AddItem start a new goroutine to execute the OnItem handler chain
func (c *Task) AddItem(i interface{}) {
	c.s.addItem(&Item{
		Ctx:  c,
		Data: i,
	})
}

// IsDownloaded return true if the request is downloaded
func (c *Task) IsDownloaded() bool {
	return c.Response != nil
}

func (c *Task) Println(v ...interface{}) { // TODO
	Logger.Print(v...)
}

func (c *Task) Error() *zerolog.Event {
	return Logger.Error()
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
