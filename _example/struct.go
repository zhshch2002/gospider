package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/zhshch2002/goreq"
	"github.com/zhshch2002/gospider"
)

func crawl() {
	s := gospider.NewSpider() // create spider

	s.AddRootTask(
		goreq.Get("https://example.com/"),
		func(t *gospider.Task) {
			h, _ := t.HTML()
			h.Find("div.a a[href]").Each(func(i int, sel *goquery.Selection) {
				t.AddTask(
					goreq.Get(sel.AttrOr("href", "")),
					func(t *gospider.Task) {
						h, _ := t.HTML()
						h.Find("div.b a[href]").Each(func(i int, sel *goquery.Selection) {
							t.AddTask(
								goreq.Get(sel.AttrOr("href", "")),
								func(t *gospider.Task) {
									h, _ := t.HTML()
									h.Find("div.a a[href]").Each(func(i int, sel *goquery.Selection) {
										t.AddTask(
											goreq.Get(sel.AttrOr("href", "")),
											func(t *gospider.Task) {

											},
										)
									})
								},
							)
						})
					},
				)
			})
		})
}

type MySpider struct{}

func (m *MySpider) handler3(t *gospider.Task) {
	t.Println("hello", t.Req.URL)
}

func (m *MySpider) handler2(t *gospider.Task) {
	h, _ := t.HTML()
	h.Find("div.b a[href]").Each(func(i int, sel *goquery.Selection) {
		t.AddTask(
			goreq.Get(sel.AttrOr("href", "")),
			m.handler3,
		)
	},
	)
}

func (m *MySpider) handler1(t *gospider.Task) {
	h, _ := t.HTML()
	h.Find("div.b a[href]").Each(func(i int, sel *goquery.Selection) {
		t.AddTask(
			goreq.Get(sel.AttrOr("href", "")),
			m.handler2,
		)
	},
	)
}

func (m *MySpider) start() {
	s := gospider.NewSpider() // create spider

	s.AddRootTask(
		goreq.Get("https://example.com/"),
		m.handler1,
	)
}
