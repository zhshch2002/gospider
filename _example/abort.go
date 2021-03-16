package main

import (
	"github.com/zhshch2002/goreq"
	"github.com/zhshch2002/gospider"
)

func main() {
	s := gospider.NewSpider() // create spider

	s.OnResp(func(t *gospider.Task) {
		t.Println("yep") // this will working as first handler
	})

	s.AddRootTask(
		goreq.Get("https://httpbin.org/get"),
		func(t *gospider.Task) {
			t.AddItem(t.Text) // this is second handler
			t.Abort()         // abort handler pipeline here
		},
		func(t *gospider.Task) {
			t.Println("this wont be print")
		},
	)

	s.Wait()
}
