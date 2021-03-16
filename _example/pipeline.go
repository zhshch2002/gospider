package main

import (
	"github.com/zhshch2002/goreq"
	"github.com/zhshch2002/gospider"
)

func main() {
	s := gospider.NewSpider() // create spider

	s.OnItem(func(t *gospider.Task, i interface{}) interface{} {
		t.Println(i) // this is working
		return "yes!!"
	})

	s.OnItem(func(t *gospider.Task, i interface{}) interface{} {
		t.Println(i) // this is working
		return nil
	})

	s.OnItem(func(t *gospider.Task, i interface{}) interface{} {
		// this on item will never be called because last OnItem return nil
		t.Println(i)
		return i
	})

	s.AddRootTask(
		goreq.Get("https://httpbin.org/get"),
		func(t *gospider.Task) { // this callback will only handle this request
			t.AddItem(t.Text) // submit result into OnItem pipeline
		},
	)

	s.Wait()
}
