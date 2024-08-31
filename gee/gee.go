package gee

import (
	"net/http"
	"strings"
	"text/template"
)

// Engine实现ServeHTTP方法，实现http.Handler接口，所有的HTTP请求，就都交给该实例处理
type Engine struct {
	*Group        // 嵌套一个Group，将Engine作为最顶层的分组，也就是说Engine拥有RouterGroup所有的能力
	router        *router
	groups        []*Group
	htmlTemplates *template.Template // 用于将模板加载进内存，
	funcMap       template.FuncMap   // 自定义模板渲染函数。
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.Group = &Group{engine: engine}
	engine.groups = []*Group{engine.Group} // 初始化是只有自己这个最顶层的分组
	return engine
}

func (engine *Engine) addRouter(method string, pattern string, handler HandlerFunc) {
	engine.router.addRouter(method, pattern, handler)
}

// 添加GET请求的方法
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRouter("GET", pattern, handler)
}

// 添加POST请求的方法
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRouter("POST", pattern, handler)
}

// 启动一个http服务器
func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		// 如果请求url的前缀是该组的前缀，就应用这组的中间件
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = e
	e.router.handle(c)
}

// 自定义渲染函数
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// 自定义加载模板
func (e *Engine) LoadHTMLGlob(pattern string) {
	// Must是一个辅助函数，它封装对返回(*Template, error)的函数的调用，并在错误非nil时panic。它旨在用于变量初始化，
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}
