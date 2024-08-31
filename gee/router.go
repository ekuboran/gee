package gee

import (
	"log"
	"net/http"
	"strings"
)

// 定义请求handler
type HandlerFunc func(*Context)

// // Router实现ServeHTTP方法，实现http.Handler接口，所有的HTTP请求，就都交给该实例处理
type router struct {
	roots    map[string]*node       // 不同请求方式的Trie 树根节点
	handlers map[string]HandlerFunc // 不同请求方式+不同路径的 HandlerFunc，key例："GET-/hello/:name"
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc)}
}

// 解析路由规则，例：将/p/:lang/doc的pattern解析[p, :lang, doc]
func (r *router) parsePattern(pattern string) (parts []string) {
	vs := strings.Split(pattern, "/")
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// 添加针对不同请求的handler
func (r *router) addRouter(method string, pattern string, handler HandlerFunc) {
	log.Printf("Router %4s - %s", method, pattern)
	parts := r.parsePattern(pattern)
	key := method + "-" + pattern
	// 没有对应请求方式的路由树，就创建一个
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &node{}
	}
	// 给路由树添加路由节点
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// 获取该请求方式+该路径的路由节点和参数
func (r *router) getRouter(method string, path string) (*node, map[string]string) {
	searchParts := r.parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts, 0)
	if n != nil {
		parts := r.parsePattern(n.pattern)
		// 解析:和*两种通配符的参数
		for index, part := range parts {
			// 例 /p/go/doc匹配到/p/:lang/doc，解析结果为：{lang: "go"}
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			// 例 /static/css/geektutu.css匹配到/static/*filepath，解析结果为{filepath: "css/geektutu.css"}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

// 获取某种请求方式的前缀树的所有路由节点
func (r *router) getRoutes(method string) []*node {
	root, ok := r.roots[method]
	if !ok {
		return nil
	}
	nodes := make([]*node, 0)
	root.travel(&nodes)
	return nodes
}

// 针对不同的方法-路径，进行不同handler处理
func (r *router) handle(c *Context) {
	node, params := r.getRouter(c.Method, c.Path)
	if node != nil {
		// 在调用匹配到的handler前，将解析出来的路由参数赋值给了c.Params
		c.Params = params
		key := c.Method + "-" + node.pattern
		// 将路由匹配得到的 Handler 添加到 c.handlers列表中（此时这个c中为[middlewares1,middlewares2,...,HandlerFunc]）
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
