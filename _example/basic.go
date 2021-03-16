package main

import (
	"github.com/zhshch2002/goreq"
	"github.com/zhshch2002/gospider"
)

func main() {
	s := gospider.NewSpider() // create spider

	s.OnResp(func(t *gospider.Task) {
		t.Println("this callback will process all response")
	})

	s.OnItem(func(t *gospider.Task, i interface{}) interface{} { // collect and save crawl result
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
