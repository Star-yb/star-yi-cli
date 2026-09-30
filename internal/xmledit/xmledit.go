// Package xmledit 在不重排原文件的前提下增删 XML 节点。
//
// 读入时保留全部空白和注释；插入新节点时沿用兄弟节点的缩进。
package xmledit

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
)

// Doc 一个可以原样写回的 XML 文件。
type Doc struct {
	*etree.Document
	path    string
	crlf    bool
	rootTag []byte
}

// Read 读取 XML 文件。
func Read(path string) (*Doc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(path, data)
}

// Parse 从内存解析，path 只用于 Save。
func Parse(path string, data []byte) (*Doc, error) {
	d := etree.NewDocument()
	d.WriteSettings.CanonicalText = true
	d.WriteSettings.CanonicalAttrVal = true
	if err := d.ReadFromBytes(data); err != nil {
		return nil, fmt.Errorf("解析 %s: %w", path, err)
	}
	if d.Root() == nil {
		return nil, fmt.Errorf("%s 没有根节点", path)
	}
	norm := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	s, e := rootStartTag(norm)
	var tag []byte
	if s >= 0 {
		tag = append([]byte(nil), norm[s:e]...)
	}
	return &Doc{Document: d, path: path, crlf: bytes.Contains(data, []byte("\r\n")), rootTag: tag}, nil
}

// rootStartTag 根元素开始标签在文本中的位置：跳过声明、注释和 DOCTYPE。
func rootStartTag(b []byte) (int, int) {
	for i := 0; i < len(b); i++ {
		if b[i] != '<' {
			continue
		}
		if i+1 < len(b) && (b[i+1] == '?' || b[i+1] == '!') {
			end := "?>"
			if bytes.HasPrefix(b[i:], []byte("<!--")) {
				end = "-->"
			} else if b[i+1] == '!' {
				end = ">"
			}
			j := bytes.Index(b[i:], []byte(end))
			if j < 0 {
				return -1, -1
			}
			i += j + len(end) - 1
			continue
		}
		var quote byte
		for j := i + 1; j < len(b); j++ {
			switch c := b[j]; {
			case quote != 0:
				if c == quote {
					quote = 0
				}
			case c == '"' || c == '\'':
				quote = c
			case c == '>':
				return i, j + 1
			}
		}
		return -1, -1
	}
	return -1, -1
}

// Bytes 序列化，保持原文件的换行风格和根标签写法。
func (d *Doc) Bytes() ([]byte, error) {
	out, err := d.WriteToBytes()
	if err != nil {
		return nil, err
	}
	out = bytes.ReplaceAll(out, []byte("\r\n"), []byte("\n"))
	if d.rootTag != nil {
		if s, e := rootStartTag(out); s >= 0 && !bytes.HasSuffix(out[s:e], []byte("/>")) && !bytes.HasSuffix(d.rootTag, []byte("/>")) {
			out = append(out[:s:s], append(append([]byte(nil), d.rootTag...), out[e:]...)...)
		}
	}
	if d.crlf {
		out = bytes.ReplaceAll(out, []byte("\n"), []byte("\r\n"))
	}
	return out, nil
}

// Save 写回原路径。
func (d *Doc) Save() error {
	out, err := d.Bytes()
	if err != nil {
		return err
	}
	return os.WriteFile(d.path, out, 0o644)
}

// Text 子元素的文本，去掉首尾空白。
func Text(e *etree.Element, tag string) string {
	if e == nil {
		return ""
	}
	c := e.SelectElement(tag)
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.Text())
}

// SetText 子元素存在时改写它的文本，返回是否改动。
func SetText(e *etree.Element, tag, value string) bool {
	c := e.SelectElement(tag)
	if c == nil {
		return false
	}
	if strings.TrimSpace(c.Text()) == value {
		return false
	}
	c.SetText(value)
	return true
}

// Escape 转义放进 XML 文本的值。
func Escape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// indentBefore 元素所在行的缩进。
func indentBefore(e *etree.Element) string {
	p := e.Parent()
	if p == nil {
		return ""
	}
	i := e.Index()
	if i <= 0 {
		return ""
	}
	cd, ok := p.Child[i-1].(*etree.CharData)
	if !ok || !cd.IsWhitespace() {
		return ""
	}
	s := cd.Data
	if j := strings.LastIndex(s, "\n"); j >= 0 {
		return s[j+1:]
	}
	return ""
}

// unit 文档的缩进单位，取根元素第一个子元素的缩进。
func unit(e *etree.Element) string {
	root := e
	for root.Parent() != nil && root.Parent().Parent() != nil {
		root = root.Parent()
	}
	for _, c := range root.ChildElements() {
		if s := indentBefore(c); s != "" {
			return s
		}
	}
	return "    "
}

func childIndent(parent *etree.Element) (child, closing string) {
	closing = indentBefore(parent)
	if kids := parent.ChildElements(); len(kids) > 0 {
		if s := indentBefore(kids[0]); s != "" {
			return s, closing
		}
	}
	return closing + unit(parent), closing
}

// build 解析片段并按目标缩进重排片段内部。片段内部用 4 个空格表示一级。
func build(fragment, indent, u string) (*etree.Element, error) {
	lines := strings.Split(strings.TrimSpace(strings.ReplaceAll(fragment, "\r\n", "\n")), "\n")
	for i := 1; i < len(lines); i++ {
		l := lines[i]
		n := 0
		for strings.HasPrefix(l, "    ") {
			l = l[4:]
			n++
		}
		lines[i] = indent + strings.Repeat(u, n) + l
	}
	d := etree.NewDocument()
	if err := d.ReadFromString(strings.Join(lines, "\n")); err != nil {
		return nil, fmt.Errorf("XML 片段不合法: %w", err)
	}
	root := d.Root()
	if root == nil {
		return nil, fmt.Errorf("XML 片段为空")
	}
	return root, nil
}

// Append 把片段加到 parent 里。after 不为空时放在它后面，否则放在最后。
func Append(parent *etree.Element, fragment string, after *etree.Element) (*etree.Element, error) {
	indent, closing := childIndent(parent)
	el, err := build(fragment, indent, unit(parent))
	if err != nil {
		return nil, err
	}
	ws := func(s string) *etree.CharData { return etree.NewText(s) }

	if after != nil && after.Parent() == parent {
		i := after.Index() + 1
		parent.InsertChildAt(i, ws("\n"+indent))
		parent.InsertChildAt(i+1, el)
		return el, nil
	}
	if len(parent.ChildElements()) == 0 {
		for i := len(parent.Child) - 1; i >= 0; i-- {
			if cd, ok := parent.Child[i].(*etree.CharData); ok && cd.IsWhitespace() {
				parent.RemoveChildAt(i)
			}
		}
		parent.AddChild(ws("\n" + indent))
		parent.AddChild(el)
		parent.AddChild(ws("\n" + closing))
		return el, nil
	}
	last := len(parent.Child) - 1
	if cd, ok := parent.Child[last].(*etree.CharData); ok && cd.IsWhitespace() {
		parent.InsertChildAt(last, ws("\n"+indent))
		parent.InsertChildAt(last+1, el)
		return el, nil
	}
	parent.AddChild(ws("\n" + indent))
	parent.AddChild(el)
	parent.AddChild(ws("\n" + closing))
	return el, nil
}

// Ensure 取子元素，没有就按缩进新建一个空元素。
func Ensure(parent *etree.Element, tag string) (*etree.Element, error) {
	if c := parent.SelectElement(tag); c != nil {
		return c, nil
	}
	return Append(parent, "<"+tag+"/>", nil)
}

// Remove 删除元素，连同它前面的缩进。
func Remove(e *etree.Element) {
	p := e.Parent()
	if p == nil {
		return
	}
	i := e.Index()
	p.RemoveChildAt(i)
	if i > 0 {
		if cd, ok := p.Child[i-1].(*etree.CharData); ok && cd.IsWhitespace() {
			p.RemoveChildAt(i - 1)
		}
	}
}
