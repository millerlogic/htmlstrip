// Copyright (C) 2018 Christopher E. Miller
// MIT license, see LICENSE file.

package htmlstrip

import (
	"html"
	"io"
	"strings"
)

// Writer strips any HTML written, calls W.Write() with the plain text.
type Writer struct {
	W                                io.Writer
	gotLT, gotLTBang, gotLTBangDash, // Set to false once determined which: tag, doctype and comment.
	gotTagAttribSlash, // Detect self-closing tag.
	inOpenTag, inTagAttribs, inCloseTag,
	inComment, inEntity bool
	curEntity        []byte // content of the current entity (if inEntity) excluding &;
	curTagName       []byte // name of the current tag (ONLY if inOpenTag or inCloseTag)
	curCommentDashes byte   // If inComment, number of dashes found in a row, looking for the end.
	inScriptTag,
	inStyleTag,
	inHeadTag bool // Keep track to skip script, style and head tags.
	plainbuf   []byte
	lastWrote  byte
	inTitleTag bool
	titlebuf   []byte
	err        error
}

var _ io.Writer = &Writer{}

func isspace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

const maxTagNameLen = 256
const maxEntityLen = 32
const maxTitleLen = 2048

// GetTitle returns the last title found, or empty if not found yet.
func (p *Writer) GetTitle() string {
	if p.inTitleTag {
		return ""
	}
	return strings.TrimSpace(string(p.titlebuf))
}

func (p *Writer) Write(data []byte) (int, error) {
	if p.err != nil {
		return 0, p.err
	}
	plainbuf := p.plainbuf[:0]
	appendPlain := func(elems ...byte) {
		if len(elems) > 0 {
			if !p.inScriptTag && !p.inStyleTag && !p.inHeadTag {
				plainbuf = append(plainbuf, elems...)
			} else if p.inTitleTag {
				if len(p.titlebuf) < maxTitleLen {
					p.titlebuf = append(p.titlebuf, elems...)
				}
				if len(p.titlebuf) > maxTitleLen {
					// TODO: don't clip in middle of UTF-8 sequence!
					//p.titlebuf = p.titlebuf[:maxTitleLen]
					p.titlebuf = p.titlebuf[:maxTitleLen-3]
					p.titlebuf = append(p.titlebuf, 0xE2, 0x80, 0xA6) // UTF-8 ellipsis.
				}
			}
			p.lastWrote = elems[len(elems)-1]
		}
	}
	gotTag := func(tagName []byte, isOpen bool, isClose bool) {
		//fmt.Printf("gotTag `%s` isOpen=%v isClose=%v\n", tagName, isOpen, isClose)
		switch strings.ToLower(string(p.curTagName)) {
		case "style":
			p.inStyleTag = isOpen && !isClose
		case "script":
			p.inScriptTag = isOpen && !isClose
		case "head":
			p.inHeadTag = isOpen && !isClose
		case "title":
			if p.inHeadTag {
				if isClose {
					p.inTitleTag = false
				}
				if isOpen && !isClose {
					p.titlebuf = nil
					p.inTitleTag = true
				}
				p.lastWrote = 0
			}
		case "br":
			if isOpen {
				appendPlain('\n')
			}
		case "div", "blockquote", "dd", "dl", "fieldset", "form",
			"h1", "h2", "h3", "h4", "h5", "h6", "hr", "noscript",
			"ol", "ul", "li", "p", "pre", "section", "table":
			if p.lastWrote != '\n' && p.lastWrote != 0 {
				appendPlain('\n')
			}
		}
	}
	for _, b := range data {
	dostate:
		switch {
		case p.gotLT:
			p.gotLT = false
			p.curTagName = p.curTagName[:0]
			if b == '!' {
				p.gotLTBang = true
			} else if b == '/' {
				p.inCloseTag = true
			} else if isspace(b) {
				appendPlain('<')
				goto dostate
			} else {
				p.inOpenTag = true
				goto dostate
			}
		case p.gotLTBang:
			p.gotLTBang = false
			if b == '-' {
				p.gotLTBangDash = true
			} else {
				p.curTagName = append(p.curTagName, '!')
				p.inOpenTag = true
				goto dostate
			}
		case p.gotLTBangDash:
			p.gotLTBangDash = false
			if b == '-' {
				p.inComment = true
				p.curCommentDashes = 2
			} else {
				p.curTagName = append(p.curTagName, '!', '-')
				p.inOpenTag = true
				goto dostate
			}
		case p.inComment:
			if b == '-' {
				p.curCommentDashes++
				if p.curCommentDashes > 2 {
					p.curCommentDashes = 2
				}
			} else if b == '>' {
				if p.curCommentDashes == 2 {
					p.inComment = false
				}
			} else {
				p.curCommentDashes = 0
			}
		case p.inOpenTag:
			switch b {
			case '<':
				// Bad HTML, let them jump into a new tag.
				appendPlain('<')
				appendPlain(p.curTagName...)
				p.inOpenTag = false
				p.gotLT = true
			case '>':
				p.inOpenTag = false
				gotTag(p.curTagName, true, false)
			case ' ', '\t', '\r', '\n', '/':
				p.inOpenTag = false
				p.inTagAttribs = true
				goto dostate // Needed for self-closing tag.
			default:
				if len(p.curTagName) < maxTagNameLen {
					p.curTagName = append(p.curTagName, b)
				}
			}
		case p.inTagAttribs:
			if b == '/' {
				p.gotTagAttribSlash = true
			} else if b == '>' {
				// If p.gotTagAttribSlash it's a self-closing tag.
				gotTag(p.curTagName, true, p.gotTagAttribSlash)
				p.inTagAttribs = false
				p.gotTagAttribSlash = false
			} else {
				if !isspace(b) {
					p.gotTagAttribSlash = false
				}
			}
		case p.inCloseTag:
			if b == '>' {
				p.inCloseTag = false
				gotTag(p.curTagName, false, true)
			} else {
				if !isspace(b) {
					if len(p.curTagName) < maxTagNameLen {
						p.curTagName = append(p.curTagName, b)
					}
				}
			}
		case p.inEntity:
			if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '#' {
				if len(p.curEntity) >= maxEntityLen {
					// Entity too long, it won't turn into anything so treat it all as text...
					p.inEntity = false
					appendPlain('&')
					appendPlain(p.curEntity...)
					goto dostate
				} else {
					p.curEntity = append(p.curEntity, b)
				}
			} else if b == ';' {
				p.inEntity = false
				appendPlain([]byte(html.UnescapeString("&" + string(p.curEntity) + ";"))...)
			} else {
				// Got invalid entity stuff, so treat it all as text...
				p.inEntity = false
				appendPlain('&')
				appendPlain(p.curEntity...)
				goto dostate
			}
		default:
			switch b {
			case '<':
				p.gotLT = true
			case '&':
				p.inEntity = true
				p.curEntity = p.curEntity[:0]
			default:
				if isspace(b) {
					if !isspace(p.lastWrote) && p.lastWrote != 0 {
						appendPlain(' ')
					}
				} else {
					appendPlain(b)
				}
			}
		}
	}
	p.plainbuf = plainbuf[:0]
	_, p.err = p.W.Write(plainbuf)
	return len(data), p.err
}
