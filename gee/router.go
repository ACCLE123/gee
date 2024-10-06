package gee

import (
	"net/http"
	"strings"
)

type router struct {
	handlers map[string]HandleFunc
	roots    map[string]*node
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]HandleFunc),
		roots:    make(map[string]*node),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
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

func (r *router) addRoute(method string, pattern string, handler HandleFunc) {
	parts := parsePattern(pattern)
	//log.Println("parts:", parts)

	key := method + "-" + pattern
	r.handlers[key] = handler

	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)

	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)

		//log.Println("path: ", searchParts)
		//log.Println("n.pattern: ", parts)

		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(parts) > 1 {
				params[part[1:]] = strings.Join(searchParts[1:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	//log.Println("handle: ", c.Method, c.Path, n, params)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		//log.Println("")
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 page not found: %s\n", c.Path)
	}
}
