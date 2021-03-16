# Gospider

[![gopkg](https://img.shields.io/badge/go-pkg-blue)](https://pkg.go.dev/github.com/zhshch2002/gospider)
[![goproxycn](https://goproxy.cn/stats/github.com/zhshch2002/gospider/badges/download-count.svg)](https://goproxy.cn)
![Go Test](https://github.com/zhshch2002/gospider/workflows/Go%20Test/badge.svg)
[![codecov](https://codecov.io/gh/zhshch2002/gospider/branch/master/graph/badge.svg)](https://codecov.io/gh/zhshch2002/gospider)

[Gospider 使用文档](https://wiki.xzhsh.ch/gospider/)

- Gospider - https://github.com/zhshch2002/gospider
- Goreq - https://github.com/zhshch2002/goreq

`Gospider`是一个轻量友好的的Go爬虫框架。

`Gospider`在管理网络请求方面使用了`Goreq`。 **‌这样分割项目使功能划分更加明确，Gospider负责管理调度任务，Goreq负责处理网络请求。** 在`Gospider`中的`goreq.Request`、`goreq.Response`和`goreq.Client`由`Goreq`提供。

## 🚀 Feature

- **优雅的 API**
- **便于组织具有复杂层级和逻辑的代码**
- **友善的分布式支持**
- **一些细节** 相对链接自动转换、字符编码自动解码、HTML/JSON 自动解析
- **丰富的扩展支持** 自动去重、失败重试、记录异常请求、控制延时、随机延时、并发、速率、Robots.txt 支持、随机 UA
- **轻量** 适于学习或快速开箱搭建

## ⚡ 网络请求

```
go get -u github.com/zhshch2002/goreq
```

`Gospider`依赖`Goreq`描述、完成网络请求，这是一个`Goreq`的简单演示，如需更多资料请查阅[Goreq GitHub repo](https://github.com/zhshch2002/goreq)或者[使用文档](https://wiki.xzhsh.ch/goreq/)。

```go
fmt.Println(goreq.Get("https://httpbin.org/get").AddParam("A","a").Do().Txt())
```

结果是：

```json
{
  "args": {
    "A": "a"
  }, 
  "headers": {
    "Accept-Encoding": "gzip", 
    "Host": "httpbin.org", 
    "User-Agent": "Go-http-client/2.0", 
    "X-Amzn-Trace-Id": "Root=1-6017ae9d-109027b5452abdd849d0161b"
  }, 
  "origin": "221.219.65.152", 
  "url": "https://httpbin.org/get?A=a"
}
```

此外：

- `resp.Resp() (*Response, error)` 获取响应本身以及网络请求错误。
- `resp.Txt() (string, error)` 自动处理完编码并解析为文本后的内容以及网络请求错误。
- `resp.HTML() (*goquery.Document, error)`解析为HTML
- `resp.XML() (*xmlpath.Node, error)`解析为XML
- `resp.BindXML(i interface{}) error`将XML绑定到struct
- `resp.JSON() (gjson.Result, error)`解析为JSON
- `resp.BindJSON(i interface{}) error`将Json绑定到struct
- `resp.Error() error` 网络请求错误。（正常情况下为`nil`）

`Goreq`可以设置中间件、更换Http Client。请见[Goreq 使用文档](https://wiki.xzhsh.ch/goreq/)。

## ⚡ 快速开始

```shell
go get -u github.com/zhshch2002/gospider
```

第一个例子：

```go
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

```

这是一个简单的爬虫，向`https://httpbin.org/get`发送请求并将结果作为`Item`存入`Spider`，`Gospider`会异步处理`OnItem`结果，不阻塞爬虫进程。

