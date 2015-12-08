/*
   Copyright 2014 Hariharan Srinath

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

/*
Package hashtag implements extraction of Twitter type hashtags, mentions and
replies form text in Go. This package partially ports extraction routines from
Twitter's official Java package at https://github.com/twitter/twitter-text
to Go and runs most of the standard twitter-text conformance tests. It does not
yet implement URL extraction (and hence URL/Hashtag overlaps), cashtags and lists

Since the package attempts to closely follow the Twitter-Text Java API, function
names may be longer than typical Go package function names
*/
package hashtag

import (
	"regexp"
	"strings"
)

const (
	unicodeSpaces = "[" +
		"\\x{0009}-\\x{000d}" + //  # White_Space # Cc   [5] <control-0009>..<control-000D>
		"\\x{0020}" + // White_Space # Zs       SPACE
		"\\x{0085}" + // White_Space # Cc       <control-0085>
		"\\x{00a0}" + // White_Space # Zs       NO-BREAK SPACE
		"\\x{1680}" + // White_Space # Zs       OGHAM SPACE MARK
		"\\x{180E}" + // White_Space # Zs       MONGOLIAN VOWEL SEPARATOR
		"\\x{2000}-\\x{200a}" + // # White_Space # Zs  [11] EN QUAD..HAIR SPACE
		"\\x{2028}" + // White_Space # Zl       LINE SEPARATOR
		"\\x{2029}" + // White_Space # Zp       PARAGRAPH SEPARATOR
		"\\x{202F}" + // White_Space # Zs       NARROW NO-BREAK SPACE
		"\\x{205F}" + // White_Space # Zs       MEDIUM MATHEMATICAL SPACE
		"\\x{3000}" + // White_Space # Zs       IDEOGRAPHIC SPACE
		"]"

	hashtagLetters      = "\\pL\\pM"
	hashtagNumerals     = "\\p{Nd}"
	hashtagSpecialChars = "/" + "\\." + "_" + "\\-" + "\\:" + //underscore
		"\\x{200c}" + // ZERO WIDTH NON-JOINER (ZWNJ)
		"\\x{200d}" + // ZERO WIDTH JOINER (ZWJ)
		"\\x{a67e}" + // CYRILLIC KAVYKA
		"\\x{05be}" + // HEBREW PUNCTUATION MAQAF
		"\\x{05f3}" + // HEBREW PUNCTUATION GERESH
		"\\x{05f4}" + // HEBREW PUNCTUATION GERSHAYIM
		"\\x{309b}" + // KATAKANA-HIRAGANA VOICED SOUND MARK
		"\\x{309c}" + // KATAKANA-HIRAGANA SEMI-VOICED SOUND MARK
		"\\x{30a0}" + // KATAKANA-HIRAGANA DOUBLE HYPHEN
		"\\x{30fb}" + // KATAKANA MIDDLE DOT
		"\\x{3003}" + // DITTO MARK
		"\\x{0f0b}" + // TIBETAN MARK INTERSYLLABIC TSHEG
		"\\x{0f0c}" + // TIBETAN MARK DELIMITER TSHEG BSTAR
		"\\x{0f0d}" // TIBETAN MARK SHAD

	hashtagLettersNumerals    = hashtagLetters + hashtagNumerals + hashtagSpecialChars
	hashtagLettersNumeralsSet = "[" + hashtagLettersNumerals + "]"
	hashtagLettersSet         = "[:" + hashtagLetters + "]"

	atSignsChars = "@\\x{FF20}"
	atSigns      = "[" + atSignsChars + "]"

	latinAccentsChars = "\\x{00c0}-\\x{00d6}\\x{00d8}-\\x{00f6}\\x{00f8}-\\x{00ff}" + // Latin-1
		"\\x{0100}-\\x{024f}" + // Latin Extended A and B
		"\\x{0253}\\x{0254}\\x{0256}\\x{0257}\\x{0259}\\x{025b}\\x{0263}\\x{0268}\\x{026f}\\x{0272}\\x{0289}\\x{028b}" + // IPA Extensions
		"\\x{02bb}" + // Hawaiian
		"\\x{0300}-\\x{036f}" + // Combining diacritics
		"\\x{1e00}-\\x{1eff}" // Latin Extended Additional (mostly for Vietnamese)
)

var validMention = regexp.MustCompile("([^A-Za-z0-9_!#$%&*" + atSignsChars + "]|^|[Rr][tT]:?)(" + atSigns + "+)([A-Za-z0-9_\\.\\-]{1,20})")

var invalidMentionMatchEnd = regexp.MustCompile("^(?:[" + atSignsChars + latinAccentsChars + "]|://)")

var validHashtag = regexp.MustCompile("(?m)(?:^|[^&" + hashtagLettersNumerals + "])(?:#|\\x{FF03})(" +
	hashtagLettersNumeralsSet + "*" + hashtagLettersSet + hashtagLettersNumeralsSet +
	"*)")

var invalidHashtagMatchEnd = regexp.MustCompile("^(?:[#＃]|://)")

var validReply = regexp.MustCompile("^(?:" + unicodeSpaces + ")*" + atSigns + "([A-Za-z0-9_]{1,20})")

/*
Entity is used by ExtractXXXXWithIndices functions to return the position
and text extracted. This may be expanded in the future to support List slugs
*/
type Entity struct {
	Start int
	End   int
	Value string
}

/*
ExtractHashtags extracts hashtags without the hash markers from input
text and returns them as a slice of strings.
*/
func ExtractHashtags(text string) []string {
	entities := ExtractHashtagsWithIndices(text)
	ret := make([]string, len(entities))

	for j, entity := range entities {
		ret[j] = entity.Value
	}
	return ret
}

/*
ExtractHashtagsWithIndices extracts hashtags without the hash markers from
input text and returns them as a slice of Entities containing start/end positions.
*/
func ExtractHashtagsWithIndices(text string) []Entity {
	if len(text) == 0 || !strings.ContainsAny(text, "#\uFF03") {
		return []Entity{}
	}

	matches := validHashtag.FindAllStringSubmatchIndex(text, -1)
	entities := []Entity{}

	for _, match := range matches {
		if !invalidHashtagMatchEnd.MatchString(text[match[1]:]) {
			value := text[match[2]:match[3]]
			if strings.Contains(value, "://") {
				continue
			}
			entities = append(entities, Entity{
				Start: match[2],
				End:   match[3],
				Value: value,
			})
		}
	}
	return entities
}

/*
ExtractMentionsWithIndices extracts mentions without the @ markers from
input text and returns them as a slice of Entities containing start/end positions.
*/
func ExtractMentionsWithIndices(text string) []Entity {
	if len(text) == 0 || !strings.ContainsAny(text, "@＠") {
		return []Entity{}
	}

	matches := validMention.FindAllStringSubmatchIndex(text, -1)
	entities := []Entity{}
	for _, match := range matches {
		if !invalidMentionMatchEnd.MatchString(text[match[1]:]) {
			entities = append(entities, Entity{
				Start: match[6],
				End:   match[7],
				Value: text[match[6]:match[7]],
			})
		}
	}
	return entities
}

/*
ExtractMentions extracts mentions without the @ markers from
input text and returns them as a slice of strings.
*/
func ExtractMentions(text string) []string {
	entities := ExtractMentionsWithIndices(text)
	ret := make([]string, len(entities))

	for j, ent := range entities {
		ret[j] = ent.Value
	}
	return ret
}

/*
ExtractReply extracts reply username without the
@ marker from input text and returns it as a string.
Empty string signals no reply username
*/
func ExtractReply(text string) string {
	if len(text) == 0 || !strings.ContainsAny(text, "@＠") {
		return ""
	}

	matches := validReply.FindAllStringSubmatchIndex(text, -1)
	for _, match := range matches {
		if !invalidMentionMatchEnd.MatchString(text[match[1]:]) {
			return text[match[2]:match[3]]
		}
	}
	return ""
}
