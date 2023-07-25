package gmi

import (
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func parseMarkdown(node ast.Node) string {
	s := ""
	for _, v := range node.GetChildren() {
		if v.AsContainer() != nil {
			s += parseMarkdown(v)
		} else {
			leaf := v.AsLeaf()
			ptr := leaf.Literal
			if ptr == nil {
				ptr = leaf.Content
			}
			if ptr == nil {
				continue
			}
			i, ok := v.GetParent().(*ast.Image)
			if ok {
				s += "=>" + string(i.Destination) + " "
			}
			l, ok := v.GetParent().(*ast.Link)
			if ok {
				s += "=>" + string(l.Destination) + " "
			}
			_, ok = v.(*ast.Text)
			if ok {
				h, ok := v.GetParent().(*ast.Heading)
				if ok {
					for i := 0; i < h.Level; i++ {
						s += "#"
					}
					s += " "
				}
				if string(ptr) == "" { continue }
				s += string(ptr) + "\n\n"
			}
			_, ok = v.(*ast.CodeBlock)
			if ok {
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
	return parseMarkdown(doc)
}
