/*
Package yaml implements a parser for a subset of the YAML data serialization
standard (http://yaml.org/). It basically only contains features that are
useful for configuration files.

Composite types

In addition to basic scalar values, the yaml package supports sequences and
mappings.

Sequences:

    - one
    - two
    - three

Mappings:

    key: value
    foo: bar

These composite types can be nested as necessary.

Sequences of sequences:

    - - one
      - two
      - three
    - - four
      - five
      - six

Mappings of mappings:

    key:
      foo: bar
    baz:
      hello: world

Sequences of mappings:

    - key:   value
      foo:   bar
    - baz:   qux
      hello: world

Mappings of sequences:

    key:
      - one
      - two
      - three
    foo:
      - four
      - five
      - six

Multiline values

In multiline values, all line feeds are converted to spaces and whitespace (not
only indenting spaces!) around each line are trimmed. So the mapping

    key: value
      goes
        here

    second:

will set "value goes here" as the value of "key".

This behavior can be overridden with literal style, where the value starts with
a single line only containing a vertical line (|). In literal style everything
is preserved, except indentation up to the first line and trailing empty lines.
So the mapping

    key: |
      value

       goes
      here

    second:

will set "value\n\n goes\nhere\n" as the value of "key".

Referenced values

The !container tag can be used for referencing values from external sources.

The tag is followed by a key whose value is read from a provided container (see
ivxv.ee/container). Usually the YAML-serialized data being parsed is included
in the same container and is referencing values that are inconvenient to
inline.

    certificate: !container cert.pem

If the key ends in ".yaml", then the referenced content is parsed as YAML and
the resulting root node is used as the value.

Comments

The number sign/hash character (#) preceded by whitespace starts a comment,
which will ignore content until the end of the line.
*/
package yaml // import "ivxv.ee/yaml"
