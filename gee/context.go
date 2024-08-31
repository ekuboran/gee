package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	// 源对象
	Req    *http.Request
	Writer http.ResponseWriter
	// 请求信息
	Path   string
	Method string
	Params map[string]string // 用于存放路由参数，如路由规则为/p/:lang/doc，路由到/p/go/doc，Params为{lang: "go"}
	// 响应信息
	StatusCode int
	// 中间件
	handlers []HandlerFunc // 中间件的定义与路由映射的 Handler 一致，处理的输入是`Context`对象
	index    int           // 记录当前执行到第几个中间件
	// engine pointer
	engine *Engine
}

func newContext(writer http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: writer,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// 获取动态路由的参数
func (c *Context) Param(key string) string {
	return c.Params[key]
}

// 获取Query参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 获取PostForm参数
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 设置请求头
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// 设置响应码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// 构造String响应（文本类型），三步：设置请求头、设置响应码、构造请求体
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// 构造Data响应(二进制)
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// 构造JSON响应
func (c *Context) Json(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// 构造HTML响应
func (c *Context) Html(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

// 依次执行其他的中间件或用户的Handler
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// 中间件功能示例函数
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.Json(code, H{"message": err})
}
