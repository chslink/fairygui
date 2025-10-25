package textutil

import (
	"strconv"
	"strings"
)

// Style describes the visual attributes for a text segment.
type Style struct {
	Color     string
	Bold      bool
	Italic    bool
	Underline bool
	Font      string
	FontSize  int
}

// Segment binds a text span with its resolved style.
type Segment struct {
	Text     string
	Style    Style
	Link     string
	ImageURL string // 图片 URL (用于 [img] 标签)
}

type stackEntry struct {
	tag      string
	style    Style
	link     string
	imageURL string
}

// ParseUBB converts a UBB formatted string into a slice of styled segments.
// Unknown tags are emitted verbatim. The base style is used for segments
// outside any explicit UBB markup.
func ParseUBB(input string, base Style) []Segment {
	if input == "" {
		return nil
	}
	stack := []stackEntry{{style: base}}
	var segments []Segment
	var builder strings.Builder

	flush := func() {
		if builder.Len() == 0 {
			return
		}
		current := builder.String()
		top := stack[len(stack)-1]
		segments = append(segments, Segment{
			Text:     current,
			Style:    top.style,
			Link:     top.link,
			ImageURL: top.imageURL,
		})
		builder.Reset()
	}

	writeLiteral := func(lit string) {
		if lit == "" {
			return
		}
		builder.WriteString(lit)
	}

	for i := 0; i < len(input); {
		ch := input[i]
		if ch != '[' {
			builder.WriteByte(ch)
			i++
			continue
		}
		end := strings.IndexByte(input[i+1:], ']')
		if end == -1 {
			builder.WriteByte(ch)
			i++
			continue
		}
		token := input[i+1 : i+1+end]
		i += end + 2
		if token == "" {
			continue
		}
		if strings.EqualFold(token, "br") {
			flush()
			top := stack[len(stack)-1]
			segments = append(segments, Segment{
				Text:     "\n",
				Style:    top.style,
				Link:     top.link,
				ImageURL: top.imageURL,
			})
			continue
		}
		if strings.EqualFold(token, "/br") {
			flush()
			continue
		}
		if token[0] == '/' {
			name := strings.ToLower(strings.TrimSpace(token[1:]))
			if name != "" {
				found := false
				for i := len(stack) - 1; i >= 1; i-- {
					if stack[i].tag == name {
						found = true
						break
					}
				}
				if !found {
					writeLiteral("[" + token + "]")
					continue
				}
			}

			// 特殊处理 img 闭合标签
			if name == "img" && builder.Len() > 0 {
				// [img]url[/img] 格式,builder 中是图片 URL
				imageURL := builder.String()
				builder.Reset()

				// 创建图片 segment
				segments = append(segments, Segment{
					Text:     "\uFFFD", // 使用 Unicode 替换字符作为图片占位符
					Style:    stack[len(stack)-1].style,
					Link:     stack[len(stack)-1].link,
					ImageURL: imageURL,
				})
			} else {
				flush()
			}

			// Pop until matching tag found.
			for len(stack) > 1 {
				entry := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if entry.tag == name || name == "" {
					break
				}
			}
			continue
		}
		name, attr := parseTag(token)
		switch name {
		case "b":
			flush()
			style := stack[len(stack)-1].style
			style.Bold = true
			stack = append(stack, stackEntry{tag: name, style: style, link: stack[len(stack)-1].link})
		case "i":
			flush()
			style := stack[len(stack)-1].style
			style.Italic = true
			stack = append(stack, stackEntry{tag: name, style: style, link: stack[len(stack)-1].link})
		case "u":
			flush()
			style := stack[len(stack)-1].style
			style.Underline = true
			stack = append(stack, stackEntry{tag: name, style: style, link: stack[len(stack)-1].link})
		case "color":
			if attr == "" {
				writeLiteral("[" + token + "]")
				continue
			}
			flush()
			style := stack[len(stack)-1].style
			style.Color = attr
			stack = append(stack, stackEntry{tag: name, style: style, link: stack[len(stack)-1].link})
		case "size":
			if attr == "" {
				writeLiteral("[" + token + "]")
				continue
			}
			if v, err := strconv.Atoi(attr); err == nil && v > 0 {
				flush()
				style := stack[len(stack)-1].style
				style.FontSize = v
				stack = append(stack, stackEntry{tag: name, style: style, link: stack[len(stack)-1].link})
			} else {
				writeLiteral("[" + token + "]")
			}
		case "font":
			if attr == "" {
				writeLiteral("[" + token + "]")
				continue
			}
			flush()
			style := stack[len(stack)-1].style
			style.Font = attr
			stack = append(stack, stackEntry{tag: name, style: style, link: stack[len(stack)-1].link})
		case "url":
			flush()
			style := stack[len(stack)-1].style
			if !style.Underline {
				style.Underline = true
			}
			stack = append(stack, stackEntry{tag: name, style: style, link: attr})
		case "img":
			// [img]url[/img] 标签用于在文本中内嵌图片
			// 图片内容在闭合标签前的文本中
			if attr == "" {
				// 没有属性,需要等待闭合标签来获取图片 URL
				flush()
				stack = append(stack, stackEntry{tag: name, style: stack[len(stack)-1].style, link: stack[len(stack)-1].link})
			} else {
				// 有属性直接使用: [img=url] 或 [img src=url]
				flush()
				// 创建特殊的图片 segment,使用特殊标记
				segments = append(segments, Segment{
					Text:     "\uFFFD", // 使用 Unicode 替换字符作为图片占位符
					Style:    stack[len(stack)-1].style,
					Link:     stack[len(stack)-1].link,
					ImageURL: attr,
				})
			}
		default:
			// Unsupported tag: emit literally.
			writeLiteral("[" + token + "]")
		}
	}
	flush()
	return segments
}

func parseTag(token string) (string, string) {
	eq := strings.IndexByte(token, '=')
	if eq == -1 {
		return strings.ToLower(strings.TrimSpace(token)), ""
	}
	name := strings.ToLower(strings.TrimSpace(token[:eq]))
	attr := strings.TrimSpace(token[eq+1:])
	if len(attr) >= 2 && ((attr[0] == '"' && attr[len(attr)-1] == '"') || (attr[0] == '\'' && attr[len(attr)-1] == '\'')) {
		attr = attr[1 : len(attr)-1]
	}
	return name, attr
}
