package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

const bin = "./testgen"

var gobin = "go"

func TestMain(m *testing.M) {
	// Prefer go binary from environment.
	if goenv, ok := os.LookupEnv("GO"); ok {
		gobin = goenv
	}

	// Generate the template source file, if not already done.
	out, err := exec.Command(gobin, "generate").CombinedOutput()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))
		fmt.Fprintln(os.Stderr, "go generate failed:", err)
		os.Exit(1)
	}

	// Build the testing binary.
	out, err = exec.Command(gobin, "build", "-o", bin).CombinedOutput()
	if err != nil {
		fmt.Fprint(os.Stderr, string(out))
		fmt.Fprintf(os.Stderr, "building %s failed: %v\n", bin, err)
		os.Exit(1)
	}

	// Run tests.
	os.Unsetenv("GOPATH") // Clear for testing.
	code := m.Run()

	// Remove the test binary.
	os.Remove(bin)
	os.Exit(code)
}

// res is the result of running the test binary.
type res struct {
	t   *testing.T
	out []byte
}

// testrun runs the test binary and assserts that it succeeds.
func testrun(t *testing.T, args ...string) *res {
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(bin, args, "failed unexpectedly:", err)
	}
	return &res{t, out}
}

// testfail runs the test binary and assserts that it fails.
func testfail(t *testing.T, args ...string) *res {
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err == nil {
		t.Logf(string(out))
		t.Fatal(bin, args, "succeeded unexpectedly")
	}
	return &res{t, out}
}

// expect checks that the result contains the pattern, reporting an error
// otherwise.
func (r *res) expect(pattern string) {
	if !grep(pattern, r.out) {
		r.t.Error("output does not contain", pattern)
	}
}

// expectnot checks that the result does NOT contain the pattern, reporting an
// error otherwise.
func (r *res) expectnot(pattern string) {
	if grep(pattern, r.out) {
		r.t.Error("output contains", pattern)
	}
}

// grep checks if b contains the pattern. Matching is performed per-line.
func grep(pattern string, b []byte) bool {
	re := regexp.MustCompile(pattern)
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if re.Match(line) {
			return true
		}
	}
	return false
}

func TestEmpty(t *testing.T) {
	r := testrun(t, "-v", "./testdata/empty")
	r.expect(`^checking \.$`)
	r.expectnot(`^generating testdata/empty/gen_types(_test)\.go$`)
}

func TestSingle(t *testing.T) {
	r := testrun(t, "-v", "./testdata/single")
	defer os.Remove("testdata/single/gen_types.go")
	r.expect(`^checking \.$`)
	r.expect(`^found Single{Field Field2} at`)
	r.expect(`^generating testdata/single/gen_types\.go$`)
	r.expectnot(`^generating testdata/single/gen_types_test\.go$`)

	// Ensure the generated code compiles.
	out, err := exec.Command(gobin, "build", "./testdata/single").CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Error("building ./testdata/single failed:", err)
	}
}

func TestWithTests(t *testing.T) {
	r := testrun(t, "-v", "./testdata/withtests")
	defer os.Remove("testdata/withtests/gen_types.go")
	defer os.Remove("testdata/withtests/gen_types_test.go")
	r.expect(`^checking \.$`)
	r.expect(`^found Main{Field} at`)
	r.expect(`^found Test{Field} at`)
	r.expect(`^generating testdata/withtests/gen_types\.go$`)
	r.expect(`^generating testdata/withtests/gen_types_test\.go$`)
}

func TestExists(t *testing.T) {
	r := testfail(t, "-v", "./testdata/exists")
	r.expect(`^checking \.$`)
	r.expect(`^found Exists{Field} at`)
	r.expect(`^error: testdata/exists/gen_types\.go already exists`)
	r.expectnot(`^generating testdata/exists/gen_types(_test)\.go$`)
}

func TestDeclared(t *testing.T) {
	r := testrun(t, "-v", "./testdata/declared")
	r.expect(`^checking \.$`)
	r.expectnot(`^found Declared{Field} at`)
	r.expectnot(`^generating testdata/declared/gen_types(_test)\.go$`)
}

func TestLocal(t *testing.T) {
	r := testrun(t, "-v", "./testdata/local")
	defer os.Remove("testdata/local/gen_types.go")
	r.expect(`^checking \.$`)
	r.expectnot(`^found InnerLocal{Field} at`)
	r.expect(`^found OuterLocal{Field} at`)
	r.expect(`^generating testdata/local/gen_types\.go$`)
	r.expectnot(`^generating testdata/local/gen_types_test\.go$`)
}

func TestDuplicate(t *testing.T) {
	r := testfail(t, "-v", "./testdata/duplicate")
	r.expect(`^checking \.$`)
	r.expect(`^found Duplicate{Field} at`)
	r.expect(`^error: .* duplicate Duplicate`)
	r.expectnot(`^generating testdata/duplicate/gen_types(_test)\.go$`)
}

func TestNoKeys(t *testing.T) {
	r := testfail(t, "-v", "./testdata/nokeys")
	r.expect(`^checking \.$`)
	r.expect(`^error: .* NoKeys must have key:value fields$`)
	r.expectnot(`^generating testdata/nokeys/gen_types(_test)\.go$`)
}

func TestUnexportedKeys(t *testing.T) {
	r := testfail(t, "-v", "./testdata/unexportedkeys")
	r.expect(`^checking \.$`)
	r.expect(`^error: .* UnexportedKeys must have exported identifier keys$`)
	r.expectnot(`^generating testdata/unexportedkeys/gen_types(_test)\.go$`)
}
