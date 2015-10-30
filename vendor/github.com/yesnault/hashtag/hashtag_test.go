package hashtag

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

type Test struct {
	Description string   `yaml:"description"`
	Text        string   `yaml:"text"`
	Expected    []string `yaml:"expected"`
}

type ExtractFunc func(string) []string

func TestConformance(t *testing.T) {
	fil, err := os.Open("tests.yml")
	if err != nil {
		t.Fatalf("Error opening tests.yml for conformance tests: %s", err)
	}
	yml, err := ioutil.ReadAll(fil)
	if err != nil {
		t.Fatalf("Error in reading from tests.yml for conformance tests :%s", err)
	}
	tests := make(map[string][]Test)
	if err := yaml.Unmarshal(yml, &tests); err != nil {
		t.Fatalf("Error in parsing YAML from tests.yml for conformance tests :%s", err)
	}

	testfuncs := map[string]ExtractFunc{
		"mentions": ExtractMentions,
		"replies":  replyTestWrapper,
		"hashtags": ExtractHashtags,
	}

	for name, testfunc := range testfuncs {
		for _, test := range tests[name] {
			got := testfunc(test.Text)
			if !reflect.DeepEqual(test.Expected, got) {
				t.Errorf("Failed %s\nText: %s\nWant: %v\nGot:%v\n", test.Description, test.Text, test.Expected, got)
			}
		}
	}
}

func replyTestWrapper(text string) []string {
	ret := ExtractReply(text)
	if ret == "" {
		return []string{}
	}
	return []string{ret}
}

func ExampleExtractHashtags() {
	fmt.Println(ExtractHashtags("this is a #hashtag but this isn't #http://example.com"))
	// Output: [hashtag]
}

func ExampleExtractHashtagsWithIndices() {
	fmt.Println(ExtractHashtagsWithIndices("this is a #hashtag but this isn't #http://example.com"))
	// Output: [{11 18 hashtag}]
}

func ExampleExtractMentions() {
	fmt.Println(ExtractMentions("mention @username1 @username2 but not user@example.com"))
	// Output: [username1 username2]
}

func ExampleExtractMentionsWithIndices() {
	fmt.Println(ExtractMentionsWithIndices("mention @username1 @username2 but not user@example.com"))
	// Output: [{9 18 username1} {20 29 username2}]
}

func ExampleExtractReply() {
	fmt.Println(ExtractReply("@username1 reply something but not to @username2"))
	// Output: username1
}
