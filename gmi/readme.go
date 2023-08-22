package gmi

import (
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func parseMarkdown(node ast.Node, isListItem bool) string {
	s := ""
	for _, v := range node.GetChildren() {
		if v.AsContainer() != nil {
			_, ok := v.(*ast.ListItem)
			_, isList := v.(*ast.List)
			s += parseMarkdown(v, ok || isListItem)
			if isList { s += "\n" }
		} else {
			leaf := v.AsLeaf()
			ptr := leaf.Literal
			if ptr == nil {
				ptr = leaf.Content
			}
			if ptr == nil {
				continue
			}
			if i, ok := v.GetParent().(*ast.Image); ok {
				s += "=>" + string(i.Destination) + " "
			}
			if l, ok := v.GetParent().(*ast.Link); ok {
				s += "=>" + string(l.Destination) + " "
			}
			if _, ok := v.(*ast.Text); ok {
				h, ok := v.GetParent().(*ast.Heading)
				if ok {
					for i := 0; i < h.Level; i++ {
						s += "#"
					}
					s += " "
				} else if isListItem {
					s += "* "
				}
				if string(ptr) == "" { continue }
				s += string(ptr) + "\n"
				if !isListItem {
					s += "\n"
				}
			}
			if _, ok := v.(*ast.CodeBlock); ok {
				s += "```\n" + string(ptr) + "\n```\n"
			}
		}
	}
	return s
}

func fromMarkdownToGmi(data string) string {
	extensions := parser.CommonExtensions
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(data))
	return parseMarkdown(doc, false)
}
