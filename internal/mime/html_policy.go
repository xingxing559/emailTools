package mimeparse

import "github.com/microcosm-cc/bluemonday"

// emailHTMLPolicy keeps layout/styles common in transactional mail (Steam, QQ, etc.)
// while still blocking scripts, forms, and embedded active content.
func newEmailHTMLPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	p.AllowElements(
		"html", "head", "body", "title", "meta", "link", "style",
		"div", "span", "p", "br", "hr",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"table", "thead", "tbody", "tfoot", "tr", "td", "th",
		"caption", "colgroup", "col",
		"ul", "ol", "li",
		"a", "img",
		"strong", "b", "em", "i", "u", "s", "sub", "sup", "small",
		"center", "font", "blockquote", "pre", "code",
		"article", "section", "header", "footer", "main",
	)

	p.AllowAttrs("style").Globally()
	p.AllowAttrs("class", "id", "role", "dir", "lang").Globally()
	p.AllowAttrs("align", "valign", "width", "height", "bgcolor", "background", "color",
		"cellpadding", "cellspacing", "border", "colspan", "rowspan").Globally()
	p.AllowAttrs("href", "target", "rel", "title", "name").OnElements("a")
	p.AllowAttrs("src", "alt", "title", "width", "height", "border").OnElements("img")
	p.AllowAttrs("content", "http-equiv", "name", "charset").OnElements("meta")
	p.AllowAttrs("type", "media").OnElements("style", "link")

	p.AllowStandardURLs()
	p.AllowDataURIImages()

	return p
}
