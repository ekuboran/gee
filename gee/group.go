package gee

import (
	"log"
	"net/http"
	"path"
)

type Group struct {
	prefix      string        // 前缀
	middlewares []HandlerFunc // 该组的中间件
	engine      *Engine       // 所有group共享一个Engine实例
}

// group创建一个子分组
func (g *Group) NewGroup(prefix string) *Group {
	engine := g.engine
	group := &Group{
		prefix: prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, group)
	return group
}

func (g *Group) addRouter(method string, comp string, handler HandlerFunc) {
	pattern := g.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	g.engine.addRouter(method, pattern, handler)
}

// GET defines the method to add GET request
func (g *Group) GET(pattern string, handler HandlerFunc) {
	g.addRouter("GET", pattern, handler)
}

// POST defines the method to add POST request
func (g *Group) POST(pattern string, handler HandlerFunc) {
	g.addRouter("POST", pattern, handler)
}

// 应用中间件
func (g *Group) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// 创建静态文件处理程序
func (g *Group) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	// StripPrefix返回一个处理程序，该处理程序通过从请求URL的路径(如果设置了RawPath)中删除给定的前缀并调用处理程序h来服务HTTP请求。
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// 检查文件是否存在
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// 暴露给开发者的方法，开发者可以将磁盘上的某个文件夹root映射到路由relativePath。
func (g *Group) Static(relativePath string, root string) {
	// Dir使用限制在特定目录树中的本机文件系统实现文件系统。
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// 注册GET处理程序
	g.GET(urlPattern, handler)
}
