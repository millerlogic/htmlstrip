package htmlstrip

import (
	"strings"
	"testing"
)

func TestHTMLTitle(t *testing.T) {
	for _, ht := range htmlTests {
		sb := &strings.Builder{}
		w := &Writer{W: sb}
		if w.GetTitle() != "" {
			panic("Unexpected title")
		}
		if _, err := w.Write([]byte(ht.html)); err != nil {
			panic(err)
		}
		hasTitle := strings.Contains(ht.html, "<title>")
		if hasTitle && w.GetTitle() == "" {
			panic("Title expected")
		} else if !hasTitle && w.GetTitle() != "" {
			panic("Unexpected title")
		}
	}

	{
		// Check for specific title.
		sb := &strings.Builder{}
		w := &Writer{W: sb}
		if w.GetTitle() != "" {
			panic("Unexpected title")
		}
		if _, err := w.Write([]byte(`<head><title> foo  bar </title></head> baz`)); err != nil {
			panic(err)
		}
		if w.GetTitle() != "foo bar" {
			panic("Title expected")
		}
	}

	{
		// Check for over-long title.
		sb := &strings.Builder{}
		w := &Writer{W: sb}
		if w.GetTitle() != "" {
			panic("Unexpected title")
		}
		fulltitle := strings.Repeat("Foo", 350) + strings.Repeat("Bar", 350)
		if _, err := w.Write([]byte(`<head><title>` + fulltitle + `</title></head> baz`)); err != nil {
			panic(err)
		}
		croptitle := fulltitle[:2048-3] + "…"
		if w.GetTitle() != croptitle {
			t.Logf("full title = `%s`", fulltitle)
			t.Logf("crop title = `%s`", croptitle)
			t.Logf(" got title = `%s`", w.GetTitle())
			panic("Over-long title is incorrect")
		}
	}
}

func TestHTMLParts(t *testing.T) {
	// isplit is a split point to break up the input html to be parsed in two parts.
	for isplit := 0; isplit < 311; isplit++ {
		for i, ht := range htmlTests {
			sb := &strings.Builder{}
			w := &Writer{W: sb}
			var err error
			if len(ht.html) >= isplit {
				inpart0 := ht.html[:isplit]
				_, err = w.Write([]byte(inpart0))
				if err != nil {
					panic(err)
				}
				inpart1 := ht.html[isplit:]
				_, err = w.Write([]byte(inpart1))
				if err != nil {
					panic(err)
				}
				expect := strings.TrimSpace(ht.plain)
				got := strings.TrimSpace(sb.String())
				if got != expect {
					t.Fatalf("htmlTests[%d] isplit=%d failed;\n input[0]: `%s`\n input[1]: `%s`\n expected: `%s`\n got: `%s`",
						i, isplit, inpart0, inpart1, expect, got)
				}
			}
		}
	}
}

type htmlTest struct {
	html, plain string
}

var htmlTests = []htmlTest{
	htmlTest{
		"foo<b>bar</b><br>f<c/>o<d />o<span id=a>bar</span>",
		"foobar\nfoobar",
	},
	htmlTest{
		"hello<!----cruel---->world",
		"helloworld",
	},
	htmlTest{
		"<!-- nothing to see here -- or here -> not even here",
		"",
	},
	htmlTest{
		"foo<!--<!-->!<!-->bar",
		"foo!bar",
	},
	htmlTest{
		"<div><p><p><script><!--</script>-->hey</script>hi",
		"hi",
	},
	htmlTest{
		"<!DOCTYPE html><html><p>foo</p><p>bar</p></html>",
		"foo\nbar",
	},
	htmlTest{
		"hello&amp;world&#x20;&#x21;",
		"hello&world !",
	},
	htmlTest{
		"<broken<tags<span>stuff",
		"<broken<tagsstuff",
	},
	htmlTest{
		"foo < bar",
		"foo < bar",
	},
	htmlTest{
		"foo&bar;baz",
		"foo&bar;baz",
	},
	htmlTest{
		"&SuperLongEntityThatWillNotTurnIntoAnythingSpecial;",
		"&SuperLongEntityThatWillNotTurnIntoAnythingSpecial;",
	},
	htmlTest{
		"&entity interruptus;",
		"&entity interruptus;",
	},
	htmlTest{
		`<!DOCTYPE html>
		<html>
		<head>
			<title>hi</title>
			<style>.foo { color: black; }</style>
		</head>
		<body>
			<p>Hi!</p>
		</body>
		</html>`,
		"Hi!",
	},
	htmlTest{
		"",
		"",
	},
}
