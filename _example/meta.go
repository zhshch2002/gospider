package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/zhshch2002/goreq"
	"github.com/zhshch2002/gospider"
)

func main() {
	s := gospider.NewSpider() // create spider

	s.AddRootTask(
		goreq.Get("https://httpbin.org/"),
		func(t *gospider.Task) {
			h, _ := t.HTML()
			t.Meta["form"] = t.Req.URL

			h.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
				t.AddTask(
					goreq.Get(sel.AttrOr("href", "")),
					func(t2 *gospider.Task) {
						t2.Println(t2.Status, t2.Req.URL, "from", t2.Meta["form"])
					},
				)
			})
		},
	)

	s.Wait()
}
