package yaml

import (
	"strings"
	"testing"

	"ivxv.ee/errors"
)

var container = map[string][]byte{
	"foo bar.txt":  []byte("foo bar"),
	"foo bar.yaml": []byte("foo: bar\nhello: world"),
}

func testParse(t *testing.T, yaml string, expected Node) {
	root, err := Parse(strings.NewReader(yaml), container)
	if err != nil {
		t.Error("syntactic analysis error:", err)
		return
	}
	if !equal(root, expected) {
		t.Errorf("yaml %q, unexpected root node: %T(%q); want %T(%q)",
			yaml, root, root, expected, expected)
	}
}

func testParseError(t *testing.T, yaml string, line, col int, cause error) {
	_, err := Parse(strings.NewReader(yaml), container)
	sae, ok := err.(SyntacticAnalysisError)
	if !ok || sae.Line.(int) != line || sae.Column.(int) != col ||
		errors.CausedBy(err, cause) == nil {

		t.Errorf("yaml %q, unexpected error: %v; want line %d, col %d, cause %T",
			yaml, err, line, col, cause)
	}
}

func TestParse(t *testing.T) {
	t.Run("reference", func(t *testing.T) {
		testParse(t, "!container foo bar.txt", scalar("foo bar"))
		testParse(t, "!container foo bar.yaml",
			mapping(map[string]Node{
				"foo":   Scalar{"foo", "bar"},
				"hello": Scalar{"hello", "world"},
			}))

		testParse(t, "foo: !container foo bar.txt",
			mapping(map[string]Node{"foo": Scalar{"foo", "foo bar"}}))

		testParse(t, "foo: !container foo bar.yaml",
			mapping(map[string]Node{
				"foo": Mapping{"foo", map[string]Node{
					"foo":   Scalar{"foo.foo", "bar"},
					"hello": Scalar{"foo.hello", "world"},
				}},
			}))

		testParseError(t, "!container ", 1, 1, new(MissingReferenceError))
		testParseError(t, "!container x.txt", 1, 1, new(InvalidContainerError))
	})

	t.Run("secondary", func(t *testing.T) {
		testParse(t, "!!str", scalar(""))
		testParse(t, "!!str 123", scalar("123"))
	})

	t.Run("indent", func(t *testing.T) {
		testParseError(t, " - first\n- second", 2, 1, new(TrailingDataError))
		testParse(t, "key:\n- value\nsuffix:",
			mapping(map[string]Node{
				"key":    Sequence{"key", []Node{Scalar{"key.0", "value"}}},
				"suffix": nil,
			}))

		testParseError(t, " key:\n  value\nkey: value", 3, 1, new(TrailingDataError))
		testParse(t, "|\n literal", scalar("literal"))
		testParse(t, "  |\n literal", scalar("literal"))
		testParseError(t, "|\nliteral", 2, 1, new(TrailingDataError))
		testParseError(t, "foo: !container\nfoo bar.txt", 1, 6, new(MissingReferenceError))
	})

	t.Run("literal", func(t *testing.T) {
		testParse(t, "|\n literal", scalar("literal"))
		testParse(t, "|\n literal\n", scalar("literal\n"))
		testParse(t, "|\n literal\n\n\n", scalar("literal\n"))
		testParse(t, "|\n hello\n\n world!\n\n foobar", scalar("hello\n\nworld!\n\nfoobar"))
		testParse(t, "|\n  hello\n   world!", scalar("hello\n world!"))
		testParse(t, "|\n õäöü: world!", scalar("õäöü: world!"))
		testParse(t, "|\n hello:\n  world", scalar("hello:\n world"))
		testParse(t, "|\n !file foo.txt", scalar("!file foo.txt"))
		testParse(t, "|\n !!str 123", scalar("!!str 123"))
	})

	t.Run("mapping", func(t *testing.T) {
		testParse(t, "key: value", mapping(map[string]Node{"key": Scalar{"key", "value"}}))
		testParse(t, "key : value", mapping(map[string]Node{"key": Scalar{"key", "value"}}))
		testParse(t, "key:\n value", mapping(map[string]Node{"key": Scalar{"key", "value"}}))
		testParse(t, "key: value\nfoo: bar",
			mapping(map[string]Node{
				"key": Scalar{"key", "value"},
				"foo": Scalar{"foo", "bar"},
			}))
		testParse(t, "key:", mapping(map[string]Node{"key": nil}))
		testParse(t, "key:\nfoo:", mapping(map[string]Node{"key": nil, "foo": nil}))

		testParseError(t, "key: value\nKey: bar", 2, 1, new(NonUniqueKeyError))
		testParseError(t, "key: |\n  literal\n trailing", 3, 2, new(MappingKeyDepthError))
		testParseError(t, ": value", 1, 1, new(NodeEmptyMappingKeyError))
		testParseError(t, "key: value\n: value", 2, 1, new(EmptyMappingKeyError))
		testParseError(t, "key: value\nfoo", 2, 1, new(EOFMissingMappingIndicatorError))
		testParseError(t, "key: value\nfoo\nbar", 3, 1, new(MappingIndicatorError))
		testParseError(t, "key: value\nkey\n: value", 3, 1, new(MappingIndicatorLineError))
	})

	t.Run("mixed", func(t *testing.T) {
		testParse(t, "literal: |\n hello\n  world",
			mapping(map[string]Node{"literal": Scalar{"literal", "hello\n world"}}))

		testParse(t, "- |\n hello\n  world",
			sequence(Scalar{"0", "hello\n world"}))

		testParse(t, "sequence:\n - first\n - second",
			mapping(map[string]Node{
				"sequence": Sequence{"sequence", []Node{
					Scalar{"sequence.0", "first"},
					Scalar{"sequence.1", "second"},
				}},
			}))

		testParse(t, "- hello: world\n  foo: bar",
			sequence(Mapping{"0", map[string]Node{
				"hello": Scalar{"0.hello", "world"},
				"foo":   Scalar{"0.foo", "bar"},
			}}))

		testParse(t, "- - first\n  - second",
			sequence(Sequence{"0", []Node{
				Scalar{"0.0", "first"},
				Scalar{"0.1", "second"},
			}}))

		testParse(t, "mapping:\n hello: world\n foo: bar",
			mapping(map[string]Node{
				"mapping": Mapping{"mapping", map[string]Node{
					"hello": Scalar{"mapping.hello", "world"},
					"foo":   Scalar{"mapping.foo", "bar"},
				}},
			}))
	})

	t.Run("scalar", func(t *testing.T) {
		testParse(t, "scalar", scalar("scalar"))
		testParse(t, "hello\nworld!", scalar("hello world!"))
		testParse(t, "\n hello\n\tworld!\r\n", scalar("hello world!"))
		testParseError(t, "scalar\nkey: value", 2, 4, new(UnexpectedMappingError))
	})

	t.Run("sequence", func(t *testing.T) {
		testParse(t, "- sequence", sequence(Scalar{"0", "sequence"}))
		testParse(t, "  - first\n  - second",
			sequence(Scalar{"0", "first"}, Scalar{"1", "second"}))

		testParseError(t, "- sequence\n: mapping", 2, 1, new(SequenceIndicatorError))
	})
}
