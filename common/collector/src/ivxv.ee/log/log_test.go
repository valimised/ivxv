package log

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2016, time.November, 3, 13, 42, 24, 0, time.UTC)

const nowstr = "2016-11-03T13:42:24Z"

// testWriter implements writer by writing into a buffer.
type testWriter struct {
	*bytes.Buffer
}

func (t testWriter) Debug(m string) error {
	_, err := fmt.Fprint(t, "debug: ", m)
	return err
}

func (t testWriter) Info(m string) error {
	_, err := fmt.Fprint(t, "info: ", m)
	return err
}

func (t testWriter) Err(m string) error {
	_, err := fmt.Fprint(t, "err: ", m)
	return err
}

func (t testWriter) Alert(m string) error {
	_, err := fmt.Fprint(t, "alert: ", m)
	return err
}

func (t testWriter) Close() error {
	return nil
}

// testContext adds a new test logger into the context.
func testContext(ctx context.Context, b *bytes.Buffer, now time.Time, cid, sid string, debug bool) context.Context {
	ctx = context.WithValue(ctx, loggerKey, &logger{
		w:   testWriter{b},
		now: func() time.Time { return now },
	})
	SetDebug(ctx, debug)
	if len(cid) > 0 {
		ctx = WithConnectionID(ctx, cid)
	}
	if len(sid) > 0 {
		ctx = WithSessionID(ctx, sid)
	}
	return ctx
}

// Message types for testing.
type testEmpty struct{}

func (t *testEmpty) EntryID() string { return "TestEmpty" }

type testString string

func (t testString) EntryID() string { return "TestString" }

type testNumber float64

func (t testNumber) EntryID() string { return "TestNumber" }

type testArray [2]string

func (t testArray) EntryID() string { return "TestArray" }

type testSlice []string

func (t testSlice) EntryID() string { return "TestSlice" }

type testBase64 []byte

func (t testBase64) EntryID() string { return "TestBase64" }

type testMap map[string]string

func (t testMap) EntryID() string { return "TestMap" }

type testStruct struct {
	Field interface{}
}

func (t testStruct) EntryID() string { return "TestStruct" }

type testStringer string

func (t testStringer) String() string { return string(t) + ".String" }

type testError string

func (t testError) Error() string { return string(t) + ".Error" }

type testNested struct {
	Nested Entry
}

func (t testNested) EntryID() string { return "TestNested" }

type testSensitive struct {
	Sensitive Sensitive
}

func (t testSensitive) EntryID() string { return "TestSensitive" }

type testErrorEntry struct {
	Err error
}

func (t testErrorEntry) EntryID() string { return "TestErrorEntry" }
func (t testErrorEntry) Cause() error    { return t.Err }
func (t testErrorEntry) Error() string   { return t.Err.Error() }

type testTime struct {
	Time time.Time
}

func (t testTime) EntryID() string { return "TestTime" }

type testCertificate struct {
	Certificate *x509.Certificate
}

func (t testCertificate) EntryID() string { return "TestCertificate" }

func TestSeverity(t *testing.T) {
	var b bytes.Buffer
	ctx := testContext(context.Background(), &b, now, "", "", true)

	tests := []struct {
		severity string
		fn       func(context.Context)
	}{
		{"debug", func(ctx context.Context) { Debug(ctx, (*testEmpty)(nil)) }},
		{"info", func(ctx context.Context) { Log(ctx, (*testEmpty)(nil)) }},
		{"err", func(ctx context.Context) { Error(ctx, testErrorEntry{testError("")}) }},
		{"alert", func(ctx context.Context) { Error(ctx, testErrorEntry{Alert(testError(""))}) }},
	}

	for _, test := range tests {
		b.Reset()
		t.Run(test.severity, func(t *testing.T) {
			test.fn(ctx)
			logged := b.String()
			if sev := logged[:strings.Index(logged, ":")]; sev != test.severity {
				t.Errorf("unexpected severity: got %q, want %q", sev, test.severity)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	var b bytes.Buffer
	var m *testEmpty

	t.Run("enabled", func(t *testing.T) {
		Debug(testContext(context.Background(), &b, now, "", "", true), m)
		if b.Len() == 0 {
			t.Error("no output when debugging enabled")
		}
	})

	b.Reset()

	t.Run("disabled", func(t *testing.T) {
		Debug(testContext(context.Background(), &b, now, "", "", false), m)
		if b.Len() > 0 {
			t.Error("unexpected output when debugging disabled")
		}
	})
}

func TestFormat(t *testing.T) {
	var b bytes.Buffer
	const prefix = `info: {"Timestamp":"` + nowstr + `",`

	// Override maxlen for testing.
	maxlen = 36

	// Parse test certificate.
	der, err := ioutil.ReadFile("testdata/certificate.der")
	if err != nil {
		t.Fatal("failed to read test certificate:", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal("failed to parse test certificate:", err)
	}

	tests := []struct {
		name   string
		cid    string
		sid    string
		entry  Entry
		output string
	}{
		{"minimal", "", "", &testEmpty{}, prefix + `"ID":"TestEmpty"}`},

		{"minimal pointer", "", "", (*testEmpty)(nil), prefix + `"ID":"TestEmpty"}`},

		{"connection ID", "cid", "", (*testEmpty)(nil),
			prefix + `"ConnectionID":"cid","ID":"TestEmpty"}`},

		{"session ID", "cid", "sid", (*testEmpty)(nil),
			prefix + `"ConnectionID":"cid","SessionID":"sid","ID":"TestEmpty"}`},

		{"string", "", "", testString("message"),
			prefix + `"ID":"TestString","Entry":"message"}`},

		{"url-encoded", "", "", testString(" \xff\"\\"),
			prefix + `"ID":"TestString","Entry":"+%FF%22%5C"}`},

		{"truncate", "", "", testString("a really long string that gets truncated"),
			prefix + `"ID":"TestString","Entry":{` +
				`"Truncated":"a+really+long+string+that+gets+trunc","FullLength":40}}`},

		{"number", "", "", testNumber(123.45),
			prefix + `"ID":"TestNumber","Entry":123.45}`},

		{"array", "", "", testArray{"message", " \xff\"\\"},
			prefix + `"ID":"TestArray","Entry":["message","+%FF%22%5C"]}`},

		{"slice", "", "", testSlice{"message", " \xff\"\\"},
			prefix + `"ID":"TestSlice","Entry":["message","+%FF%22%5C"]}`},

		{"base64", "", "", testBase64("byte slice"),
			prefix + `"ID":"TestBase64","Entry":"Ynl0ZSBzbGljZQ=="}`},

		{"sensitive", "", "", testSensitive{Sensitive: Sensitive("sensitive data")},
			prefix + `"ID":"TestSensitive","Entry":{"Sensitive":{` +
				`"Hash":"HQjFmEMNLY2Y+j0/SejcozHaqBiVjYvw2/oqo4TYp/0=","Length":14}}}`},

		{"map", "", "", testMap{" \xff\"\\": " \xff\"\\"},
			prefix + `"ID":"TestMap","Entry":{"+%FF%22%5C":"+%FF%22%5C"}}`},

		{"struct", "", "", testStruct{Field: " \xff\"\\"},
			prefix + `"ID":"TestStruct","Entry":{"Field":"+%FF%22%5C"}}`},

		{"stringer", "", "", testStruct{Field: testStringer("message")},
			prefix + `"ID":"TestStruct","Entry":{"Field":"message.String"}}`},

		{"error", "", "", testErrorEntry{Err: testError("message")},
			prefix + `"ID":"TestErrorEntry","Entry":{"Err":"message.Error"}}`},

		{"nested", "", "", testNested{Nested: testString(" \xff\"\\")},
			prefix + `"ID":"TestNested","Entry":{"Nested":{"ID":"TestString","Entry":"+%FF%22%5C"}}}`},

		{"nested empty", "", "", testNested{Nested: &testEmpty{}},
			prefix + `"ID":"TestNested","Entry":{"Nested":{"ID":"TestEmpty"}}}`},

		{"nested empty pointer", "", "", testNested{Nested: (*testEmpty)(nil)},
			prefix + `"ID":"TestNested","Entry":{"Nested":{"ID":"TestEmpty"}}}`},

		{"gen", "", "", TestGenerated{}, prefix + `"ID":"ivxv.ee/log.TestGenerated"}`},

		{"alert", "", "", testErrorEntry{Err: Alert(testErrorEntry{Err: testError("message")})},
			prefix + `"ID":"TestErrorEntry","Entry":{"Err":{` +
				`"ID":"TestErrorEntry","Entry":{"Err":"message.Error"}}}}`},

		{"time", "", "", testTime{
			Time: time.Date(2017, time.February, 28, 16, 42, 35, 00, time.UTC),
		}, prefix + `"ID":"TestTime","Entry":{"Time":"2017-02-28T16%3A42%3A35Z"}}`},

		{"time nano", "", "", testTime{
			Time: time.Date(2017, time.February, 28, 16, 42, 35, 123456789, time.UTC),
		}, prefix + `"ID":"TestTime","Entry":{"Time":"2017-02-28T16%3A42%3A35.123456789Z"}}`},

		{"certificate", "", "", testCertificate{Certificate: cert},
			prefix + `"ID":"TestCertificate","Entry":{"Certificate":{` +
				`"Subject":"CN%3DConfiguration+Signer",` +
				`"Issuer":"CN%3DConfiguration+Signer",` +
				`"Serial":"995d4b1336861e79",` +
				`"Hash":"49B434Z/brFreLtC7fTop2vIKiImMMTQwK+ArARBJU4="}}}`},
	}

	for _, test := range tests {
		b.Reset()
		t.Run(test.name, func(t *testing.T) {
			ctx := testContext(context.Background(), &b, now, test.cid, test.sid, true)
			Log(ctx, test.entry)
			if got := b.String(); got != test.output {
				t.Errorf("unexpected output,\ngot:  %s\nwant: %s", got, test.output)
			}
		})
	}
}
