package htmlstrip

import (
	"strings"
	"testing"
)

func TestHTML(t *testing.T) {
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
