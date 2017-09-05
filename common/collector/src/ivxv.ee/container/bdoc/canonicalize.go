package bdoc

import (
	"fmt"
	"regexp"
	"strings"
)

type xmlns struct {
	name string
	url  string
}

var closingTags = regexp.MustCompile(`<(\w+(:\w+)?)([^>]*)/>`)

// canonicalize takes as input a xml string and a tag name. It then extracts
// and normalizes the tag and its contents according to the specification at
// http://www.w3.org/2006/12/xml-c14n11
//
// nolint: gocyclo, this function needs to be rewritten as a whole so do not
// worry about cyclomatic complexity at the moment.
func canonicalize(whole, wanted string, nameSpaces []xmlns) ([]byte, error) {
	// Make sure all linefeeds are "\n"
	whole = strings.Replace(whole, "\r\n", "\n", -1)

	// Escape regex
	wanted = regexp.QuoteMeta(wanted)

	// Get the wanted tag including the opening and closing tag
	re := regexp.MustCompile("<" + wanted + `(\s(?s:.)*?)?(>(?s:.)*?</` + wanted + `\s*>|/>)`)
	wholeWanteds := re.FindAllString(whole, -1)
	switch len(wholeWanteds) {
	case 0:
		return nil, MatchingTagNotFoundError{Tag: wanted}
	case 1:
	default:
		return nil, MultipleMatchingTagsError{Tag: wanted, Count: len(wholeWanteds)}
	}
	wholeWanted := wholeWanteds[0]

	// Check for nested elements with the same tag.
	if len(re.FindString(wholeWanted[1:])) > 0 {
		return nil, NestedMatchingTagError{Tag: wanted}
	}

	// If the tag has a namespace-prefix, then forbid any non-prefixed tags.
	if split := strings.SplitN(wanted, ":", 2); len(split) > 1 {
		if strings.Contains(whole, "<"+split[1]) {
			return nil, NonPrefixedWantedTagError{Tag: wanted}
		}
	}

	// Put together a namespace string
	var nameSpaceString string
	for _, ns := range nameSpaces {
		nameSpaceString = nameSpaceString + fmt.Sprintf(` xmlns:%v="%v"`, ns.name, ns.url)
	}

	i := len(wanted) + 1
	wantedWithNSRaw := make([]byte, i)
	copy(wantedWithNSRaw, wholeWanted[:i])
	wantedWithNSRaw = append(wantedWithNSRaw, []byte(nameSpaceString)...)

	// Remove previous namespaces and attributes
	index := strings.Index(wholeWanted, ">")
	rawAttrs := wholeWanted[i:index]
	split := strings.Split(rawAttrs, " ")
	var attrs []string
	for _, attr := range split {
		if !strings.HasPrefix(attr, "xmlns") && attr != "" {
			attrs = append(attrs, attr)
		}
	}

	wantedWithNSRaw = append(wantedWithNSRaw, []byte(" "+strings.Join(attrs, " "))...)
	wantedWithNSRaw = append(wantedWithNSRaw, []byte(wholeWanted[index:])...)
	wantedWithNS := string(wantedWithNSRaw)

	// Attribute values are normalized
	// Removes unnecessary whitespace between attributes so that all attributes are
	// separated by one space except the end of the tag. Also sees to it that no
	// whitespace is removed from attribute values and tag contents.
	// Output: <ns:tag attr1="value1" attr2="val ue2" attr3="va  lu e3">
	var normalized []rune
	var tag bool
	check := true
	for _, val := range wantedWithNS {
		if tag {
			if val == '"' || val == '\'' {
				check = !check
			}
			if check && (val == '/' || val == '>') {
				if normalized[len(normalized)-1] == ' ' {
					normalized = normalized[:len(normalized)-1]
				}
			}
			if check && (val == ' ' || val == '\t' || val == '\n' || val == '\r') {
				val = ' '
				if normalized[len(normalized)-1] == ' ' {
					continue
				}
			}
		}
		if val == '<' || val == '>' {
			tag = !tag
		}
		normalized = append(normalized, val)
	}
	wantedWithNS = string(normalized)

	// Add missing closing tags where short tags have been used.
	// <ns:tag attr="value"/> => <ns:tag attr="value"></ns:tag>
	wantedWithNS = closingTags.ReplaceAllString(wantedWithNS, "<${1}${3}></${1}>")

	return []byte(wantedWithNS), nil
}
