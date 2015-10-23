package html

import (
	"io/ioutil"
	"testing"
)

const SAMPLE_DOC = `
<!DOCTYPE html>
<html>
	<head>
		<title>Test Document</title>
		<meta name="referrer" content="origin">
		<link rel="stylesheet" type="text/css" href="foo.css?123456">
		<script type="text/javascript">
			function hide(id) {
				var el = document.getElementById(id);
				if (el) { el.style.visibility = 'hidden'; }
			}
		</script>
	</head>
	<body>
		<div class="container">
			<h1>Hello World!</h1>
			<a href="/internal" class="my-link-class blink">Test Internal Link</a>
			<a href="http://test/external" target="_blank">Test External Link</a>
		</div>
		<div class="footer">
			<!-- XML COMMENTS BITCHES. -->
		</div>
	</body>
</html>`

const SNIPPET = `<div id="first"><br><p><h1>Test!</h1></p></div><div id="second"><h2>Test 2!</h2></div>`
const SNIPPET_INVALID = `<div id="first"><br><p><h1>Test!</p></h1></div><div id="second"><h2>Test 2!</h2></div>`

func TestParsingSnippet(t *testing.T) {
	doc, parseError := Parse(SNIPPET)
	if parseError != nil {
		t.Error(parseError.Error())
		t.FailNow()
	}

	if len(doc.Children) == 0 {
		t.Error("doc children length is 0")
		t.FailNow()
	}

	if doc.Children[0].Attributes["id"] != "first" {
		t.Errorf("Invalid first child:%s", doc.Children[0].ToString())
		t.FailNow()
	}

	if doc.Children[1].Attributes["id"] != "second" {
		t.Errorf("Invalid first child:%s", doc.Children[1].ToString())
		t.FailNow()
	}
}

func TestParsingDocument(t *testing.T) {
	doc, parseError := Parse(SAMPLE_DOC)
	if parseError != nil {
		t.Error(parseError.Error())
		t.FailNow()
	}

	if len(doc.NonTextChildren()) == 0 {
		t.Error("doc children length is 0")
		t.FailNow()
	}

	textElements := doc.GetElementsByTagName("text")
	if len(textElements) == 0 {
		t.Error("`text` element count is 0")
		t.FailNow()
	}
}

func readFileContents(filename string) string {
	reader, _ := ioutil.ReadFile(filename)
	return string(reader)
}

func TestParsingMocks(t *testing.T) {
	mock_files := []string{
		"news.ycombinator.com.html",
		"nytimes.com.html",
		"blendlabs.com.html",
	}

	for _, mock_file := range mock_files {
		corpus := readFileContents("mocks/" + mock_file)
		doc, parseError := Parse(corpus)
		if parseError != nil {
			t.Errorf("error with %s: %s", mock_file, parseError.Error())
			t.FailNow()
		}

		if len(doc.NonTextChildren()) == 0 {
			t.Error("doc children length is 0")
			t.FailNow()
		}

		textElements := doc.GetElementsByTagName("text")
		if len(textElements) == 0 {
			t.Error("`text` element count is 0")
			t.FailNow()
		}
	}
}

func TestParsingInvalid(t *testing.T) {
	_, parseError := Parse(SNIPPET_INVALID)
	if parseError == nil {
		t.Error("Should have errored.")
		t.FailNow()
	}
}

func TestElementStack(t *testing.T) {
	stack := &elementStack{}

	if stack.Count != 0 {
		t.Errorf("initial stack storage count is not 0, is: %d", stack.Count)
		t.FailNow()
	}

	stack.Push(Element{ElementName: "br", IsVoid: true})

	if stack.Count != 1 {
		t.Errorf("stack storage count is not 1, is: %d", stack.Count)
		t.FailNow()
	}

	stack.Push(Element{ElementName: "div", Attributes: map[string]string{"class": "first"}})

	if stack.Count != 2 {
		t.Errorf("stack storage count is not 2, is: %d", stack.Count)
		t.FailNow()
	}

	stack.Push(Element{ElementName: "div", Attributes: map[string]string{"class": "second"}})

	if stack.Count != 3 {
		t.Errorf("stack storage count is not 3, is: %d", stack.Count)
		t.FailNow()
	}

	stack_string := stack.ToString()
	if stack_string != "br > div > div" {
		t.Errorf("stack .ToString() invalid: %s", stack_string)
		t.FailNow()
	}

	if stack.Peek().ElementName != "div" {
		t.Error("top of stack is not a `div`")
		t.FailNow()
	}

	div := stack.Pop()
	if div.ElementName != "div" {
		t.Error("first popped element should be a div")
		t.FailNow()
	}

	if div.Attributes["class"] != "second" {
		t.Error("first popped element `class` should be `second`")
		t.FailNow()
	}

	if stack.Count != 2 {
		t.Error("stack count should be 2 after popping first element.")
		t.FailNow()
	}
}

func TestReadUntilTag(t *testing.T) {
	cursor := 0
	valid := "      this is a test of reading until the tag <area/>"

	results, results_err := readUntilTag([]rune(valid), &cursor)
	if results_err != nil {
		t.Error(results_err.Error())
		t.FailNow()
	}

	if string(results) != "      this is a test of reading until the tag " {
		t.Error("Incorrect results: '" + string(results) + "'")
		t.FailNow()
	}

	cursor = 0
	no_tag := "there is no tag."

	results, results_err = readUntilTag([]rune(no_tag), &cursor)
	if results_err != nil {
		t.Error(results_err.Error())
		t.FailNow()
	}
	if string(results) != no_tag {
		t.Error("Incorrect results.")
		t.FailNow()
	}

	cursor = 0
	only_tag := "<a href='things.html'>things</a>"
	results, results_err = readUntilTag([]rune(only_tag), &cursor)
	if results_err != nil {
		t.Error(results_err.Error())
		t.FailNow()
	}
	if string(results) != EMPTY {
		t.Error("Incorrect results.")
		t.FailNow()
	}

	cursor = 0
	starts_tag := "<br/> more text ..."
	results, results_err = readUntilTag([]rune(starts_tag), &cursor)
	if results_err != nil {
		t.Error(results_err.Error())
		t.FailNow()
	}
	if string(results) != EMPTY {
		t.Error("Incorrect results.")
		t.FailNow()
	}
}

func TestReadUntilScriptTagClose(t *testing.T) {
	test_cases := map[string]string{
		`var a = "abc";</script>`:      `var a = "abc";`,
		`alert('</script>');</script>`: `alert('</script>');`,

		`//</script>
		var foo = "bar";
		</script>`: `//</script>
		var foo = "bar";
		`,

		`var foo = 'bar';
		/* this is a block 
		comment and is annoying */
		foo = 'baz';
		</script>`: `var foo = 'bar';
		/* this is a block 
		comment and is annoying */
		foo = 'baz';
		`,
	}

	for test, expected := range test_cases {
		cursor := 0
		results, results_err := readUntilScriptTagClose([]rune(test), &cursor, "text/javascript")
		if results_err != nil {
			t.Error("error occurred.")
			t.FailNow()

		}
		if len(results) == 0 {
			t.Error("empty results.")
			t.FailNow()
		}

		if expected != string(results) {
			t.Errorf("expected: '%s' actual: '%s'", expected, string(results))
			t.FailNow()
		}
	}
}

func TestReadWhitespace(t *testing.T) {
	test_string := "     \n\t     this is a test string ..."
	cursor := 0
	results, results_err := readWhitespace([]rune(test_string), &cursor)
	if results_err != nil {
		t.Error(results_err.Error())
		t.FailNow()
	}

	if string(results) != "     \n\t     " {
		t.Error("Incorrect results.")
		t.FailNow()
	}
}

func TestReadTag(t *testing.T) {
	testCases := map[string]Element{
		"<!DOCTYPE>":                 Element{ElementName: "DOCTYPE", IsVoid: true},
		"<!DOCTYPE html>":            Element{ElementName: "DOCTYPE", IsVoid: true, Attributes: map[string]string{"html": ""}},
		"<!-- this is a comment -->": Element{ElementName: "XML COMMENT", IsVoid: true, IsComment: true, InnerHTML: " this is a comment "},
		"<br>":                                         Element{ElementName: "br", IsVoid: true},
		"<br/>":                                        Element{ElementName: "br", IsVoid: true},
		"</div>":                                       Element{ElementName: "div", IsVoid: false, IsClose: true},
		"</ div>":                                      Element{ElementName: "div", IsVoid: false, IsClose: true},
		"< /div>":                                      Element{ElementName: "div", IsVoid: false, IsClose: true},
		"<div class=\"content\">":                      Element{ElementName: "div", Attributes: map[string]string{"class": "content"}},
		"<div class=\"with='quotes'\">":                Element{ElementName: "div", Attributes: map[string]string{"class": "with='quotes'"}},
		"<div class='with=\"escaped_quotes\"'>":        Element{ElementName: "div", Attributes: map[string]string{"class": "with=\"escaped_quotes\""}},
		"<a class=\"my-link\" href=\"/test/route\" />": Element{ElementName: "a", IsVoid: true, Attributes: map[string]string{"class": "my-link", "href": "/test/route"}},
	}

	for tag, expectedResult := range testCases {
		cursor := 0
		actualResult, parseError := readTag([]rune(tag), &cursor)

		if parseError != nil {
			t.Error(parseError.Error())
			t.FailNow()
		}
		if !expectedResult.EqualTo(actualResult) {
			t.Error("Invalid parsed tag results.")
			t.Errorf("\tExpected : %s", expectedResult.ToString())
			t.Errorf("\tActual   : %s", actualResult.ToString())
			t.Fail()
		}
	}
}
