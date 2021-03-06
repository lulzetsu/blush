package blush_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/arsham/blush/blush"
	"github.com/pkg/errors"
)

func TestWriteToErrors(t *testing.T) {
	w := new(bytes.Buffer)
	e := errors.New("something")
	nn := 10
	bw := &badWriter{
		writeFunc: func([]byte) (int, error) {
			return nn, e
		},
	}
	r := ioutil.NopCloser(bytes.NewBufferString("something"))
	tcs := []struct {
		name    string
		b       *blush.Blush
		writer  io.Writer
		wantN   int
		wantErr string
	}{
		{"no input", &blush.Blush{}, w, 0, blush.ErrNoReader.Error()},
		{"no writer", &blush.Blush{Reader: r}, nil, 0, blush.ErrNoWriter.Error()},
		{"bad writer", &blush.Blush{Reader: r, NoCut: true}, bw, nn, e.Error()},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			n, err := tc.b.WriteTo(tc.writer)
			if err == nil {
				t.Error("New(): err = nil, want error")
				return
			}
			if int(n) != tc.wantN {
				t.Errorf("l.WriteTo(): n = %d, want %d", n, tc.wantN)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want `%s` in `%s`", tc.wantErr, err.Error())
			}
		})
	}
}

func TestWriteToNoMatch(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "testdata")
	r, err := blush.NewMultiReader(blush.WithPaths([]string{location}, true))
	if err != nil {
		t.Fatal(err)
	}
	b := &blush.Blush{
		Reader:  r,
		Finders: []blush.Finder{blush.NewExact("SHOULDNOTFINDTHISONE", blush.NoColour)},
	}
	buf := new(bytes.Buffer)
	n, err := b.WriteTo(buf)
	if err != nil {
		t.Errorf("err = %v, want %v", err, nil)
	}
	if n != 0 {
		t.Errorf("b.WriteTo(): n = %d, want %d", n, 0)
	}
	if buf.Len() > 0 {
		t.Errorf("buf.Len() = %d, want 0", buf.Len())
	}
}

func TestWriteToMatchNoColourPlain(t *testing.T) {
	match := "TOKEN"
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "testdata")
	r, err := blush.NewMultiReader(blush.WithPaths([]string{location}, true))
	if err != nil {
		t.Fatal(err)
	}
	b := &blush.Blush{
		Reader:  r,
		Finders: []blush.Finder{blush.NewExact(match, blush.NoColour)},
	}

	buf := new(bytes.Buffer)
	n, err := b.WriteTo(buf)
	if err != nil {
		t.Errorf("err = %v, want %v", err, nil)
	}
	if buf.Len() == 0 {
		t.Errorf("buf.Len() = %d, want > 0", buf.Len())
	}
	if int(n) != buf.Len() {
		t.Errorf("b.WriteTo(): n = %d, want %d", int(n), buf.Len())
	}
	if !strings.Contains(buf.String(), match) {
		t.Errorf("want `%s` in `%s`", match, buf.String())
	}
	if strings.Contains(buf.String(), "[38;5;") {
		t.Errorf("didn't expect colouring: `%s`", buf.String())
	}
	if strings.Contains(buf.String(), leaveMeHere) {
		t.Errorf("didn't expect to see %s", leaveMeHere)
	}
}

func TestWriteToMatchColour(t *testing.T) {
	match := blush.Colourise("TOKEN", blush.Blue)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "testdata")
	r, err := blush.NewMultiReader(blush.WithPaths([]string{location}, true))
	if err != nil {
		t.Fatal(err)
	}
	b := &blush.Blush{
		Reader:  r,
		Finders: []blush.Finder{blush.NewExact("TOKEN", blush.Blue)},
	}

	buf := new(bytes.Buffer)
	n, err := b.WriteTo(buf)
	if err != nil {
		t.Errorf("err = %v, want %v", err, nil)
	}
	if buf.Len() == 0 {
		t.Errorf("buf.Len() = %d, want > 0", buf.Len())
	}
	if int(n) != buf.Len() {
		t.Errorf("b.WriteTo(): n = %d, want %d", int(n), buf.Len())
	}
	if !strings.Contains(buf.String(), match) {
		t.Errorf("want `%s` in `%s`", match, buf.String())
	}
	if strings.Contains(buf.String(), leaveMeHere) {
		t.Errorf("didn't expect to see %s", leaveMeHere)
	}
}

func TestWriteToMatchCountColour(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name      string
		recursive bool
		count     int
	}{
		{"ONE", false, 1},
		{"ONE", true, 3 * 1},
		{"TWO", false, 2},
		{"TWO", true, 3 * 2},
		{"THREE", false, 3},
		{"THREE", true, 3 * 3},
		{"FOUR", false, 4},
		{"FOUR", true, 3 * 4},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			location := path.Join(pwd, "testdata")
			r, err := blush.NewMultiReader(blush.WithPaths([]string{location}, tc.recursive))
			if err != nil {
				t.Fatal(err)
			}

			match := blush.Colourise(tc.name, blush.Red)
			b := &blush.Blush{
				Reader:  r,
				Finders: []blush.Finder{blush.NewExact(tc.name, blush.Red)},
			}

			buf := new(bytes.Buffer)
			n, err := b.WriteTo(buf)
			if err != nil {
				t.Errorf("b.WriteTo(): err = %v, want %v", err, nil)
			}
			if int(n) != buf.Len() {
				t.Errorf("b.WriteTo(): n = %d, want %d", int(n), buf.Len())
			}
			count := strings.Count(buf.String(), match)
			if count != tc.count {
				t.Errorf("count = %d, want %d", count, tc.count)
			}
			if strings.Contains(buf.String(), leaveMeHere) {
				t.Errorf("didn't expect to see %s", leaveMeHere)
			}
		})
	}
}

func TestWriteToMultiColour(t *testing.T) {
	two := blush.Colourise("TWO", blush.Magenta)
	three := blush.Colourise("THREE", blush.Red)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "testdata")
	r, err := blush.NewMultiReader(blush.WithPaths([]string{location}, true))
	if err != nil {
		t.Fatal(err)
	}
	b := &blush.Blush{
		Reader: r,
		Finders: []blush.Finder{
			blush.NewExact("TWO", blush.Magenta),
			blush.NewExact("THREE", blush.Red),
		},
	}

	buf := new(bytes.Buffer)
	n, err := b.WriteTo(buf)
	if err != nil {
		t.Errorf("err = %v, want %v", err, nil)
	}
	if buf.Len() == 0 {
		t.Errorf("buf.Len() = %d, want > 0", buf.Len())
	}
	if int(n) != buf.Len() {
		t.Errorf("b.WriteTo(): n = %d, want %d", int(n), buf.Len())
	}
	count := strings.Count(buf.String(), two)
	if count != 2*3 {
		t.Errorf("count = %d, want %d", count, 2*3)
	}
	count = strings.Count(buf.String(), three)
	if count != 3*3 {
		t.Errorf("count = %d, want %d", count, 3*3)
	}
	if strings.Contains(buf.String(), leaveMeHere) {
		t.Errorf("didn't expect to see %s", leaveMeHere)
	}
}

func TestWriteToMultiColourColourMode(t *testing.T) {
	two := blush.Colourise("TWO", blush.Magenta)
	three := blush.Colourise("THREE", blush.Red)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	location := path.Join(pwd, "testdata")
	r, err := blush.NewMultiReader(blush.WithPaths([]string{location}, true))
	if err != nil {
		t.Fatal(err)
	}
	b := &blush.Blush{
		Reader: r,
		NoCut:  true,
		Finders: []blush.Finder{
			blush.NewExact("TWO", blush.Magenta),
			blush.NewExact("THREE", blush.Red),
		},
	}

	buf := new(bytes.Buffer)
	n, err := b.WriteTo(buf)
	if err != nil {
		t.Errorf("err = %v, want %v", err, nil)
	}
	if buf.Len() == 0 {
		t.Errorf("buf.Len() = %d, want > 0", buf.Len())
	}
	if int(n) != buf.Len() {
		t.Errorf("b.WriteTo(): n = %d, want %d", int(n), buf.Len())
	}
	count := strings.Count(buf.String(), two)
	if count != 2*3 {
		t.Errorf("count = %d, want %d", count, 2*3)
	}
	count = strings.Count(buf.String(), three)
	if count != 3*3 {
		t.Errorf("count = %d, want %d", count, 3*3)
	}
	count = strings.Count(buf.String(), leaveMeHere)
	if count != 1 {
		t.Errorf("count = %d, want to see `%s` exactly %d times", count, leaveMeHere, 1)
	}
}

func TestWriteToMultipleMatchInOneLine(t *testing.T) {
	line1 := "this is an example\n"
	line2 := "someone should find this line\n"
	input1 := bytes.NewBuffer([]byte(line1))
	input2 := bytes.NewBuffer([]byte(line2))
	r := ioutil.NopCloser(io.MultiReader(input1, input2))
	match := fmt.Sprintf(
		"someone %s find %s line",
		blush.Colourise("should", blush.Red),
		blush.Colourise("this", blush.Magenta),
	)
	out := new(bytes.Buffer)

	b := &blush.Blush{
		Reader: r,
		Finders: []blush.Finder{
			blush.NewExact("this", blush.Magenta),
			blush.NewExact("should", blush.Red),
		},
	}

	b.WriteTo(out)
	lines := strings.Split(out.String(), "\n")
	example := lines[1]
	if strings.Contains(example, "is an example") {
		example = lines[0]
	}
	if example != match {
		t.Errorf("example = %s, want %s", example, match)
	}
}

func TestBlushClosesReader(t *testing.T) {
	var called bool
	input := bytes.NewBuffer([]byte("DwgQnpvro5bVvrRwBB"))
	w := nopCloser{
		Reader: input,
		closeFunc: func() error {
			called = true
			return nil
		},
	}
	b := &blush.Blush{
		Reader: w,
	}
	err := b.Close()
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}
	if !called {
		t.Error("didn't close the reader")
	}
}

func TestBlushReadOneStream(t *testing.T) {
	input := bytes.NewBuffer([]byte("one two three four"))
	match := blush.NewExact("three", blush.Blue)
	r := ioutil.NopCloser(input)
	b := &blush.Blush{
		Finders: []blush.Finder{match},
		Reader:  r,
	}
	defer b.Close()
	emptyP := make([]byte, 10)
	tcs := []struct {
		name    string
		p       []byte
		wantErr error
		wantLen int
		wantP   string
	}{
		{"one", make([]byte, len("one ")), nil, len("one "), "one "},
		{"two", make([]byte, len("two ")), nil, len("two "), "two "},
		{"three", make([]byte, len(match.String())), nil, len(match.String()), match.String()},
		{"four", make([]byte, len(" four\n")), nil, len(" four\n"), " four\n"}, // there is always a new line after each reader.
		{"empty", emptyP, io.EOF, 0, string(emptyP)},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			n, err := b.Read(tc.p)
			if err != tc.wantErr {
				t.Error(err)
			}
			if n != tc.wantLen {
				t.Errorf("b.Read(): n = %d, want %d", n, tc.wantLen)
			}
			if string(tc.p) != tc.wantP {
				t.Errorf("p = `%s`, want `%s`", tc.p, tc.wantP)
			}
		})
	}
}

func TestBlushReadTwoStreams(t *testing.T) {
	b1 := []byte("one for all\n")
	b2 := []byte("all for one\n")
	input1 := bytes.NewBuffer(b1)
	input2 := bytes.NewBuffer(b2)
	match := blush.NewExact("one", blush.Blue)
	r := ioutil.NopCloser(io.MultiReader(input1, input2))
	b := &blush.Blush{
		Finders: []blush.Finder{match},
		Reader:  r,
	}
	defer b.Close()

	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(b)
	if err != nil {
		t.Error(err)
	}
	expectLen := len(b1) + len(b2) - len("one")*2 + len(match.String())*2
	if int(n) != expectLen {
		t.Errorf("b.Read(): n = %d, want %d", n, expectLen)
	}
	expectStr := fmt.Sprintf("%s%s",
		strings.Replace(string(b1), "one", match.String(), 1),
		strings.Replace(string(b2), "one", match.String(), 1),
	)
	if buf.String() != expectStr {
		t.Errorf("buf.String() = %s, want %s", buf.String(), expectStr)
	}
}

func TestBlushReadHalfWay(t *testing.T) {
	b1 := []byte("one for all\n")
	b2 := []byte("all for one\n")
	input1 := bytes.NewBuffer(b1)
	input2 := bytes.NewBuffer(b2)
	match := blush.NewExact("one", blush.Blue)
	r := ioutil.NopCloser(io.MultiReader(input1, input2))
	b := &blush.Blush{
		Finders: []blush.Finder{match},
		Reader:  r,
	}
	p := make([]byte, len(b1))
	_, err := b.Read(p)
	if err != nil {
		t.Error(err)
	}
	n, err := b.Read(p)
	if n != len(b1) {
		t.Errorf("b.Read(): n = %d, want %d", n, len(b1))
	}
	if err != nil {
		t.Errorf("b.Read(): err = %v, want %v", err, nil)
	}
}

func TestBlushReadOnClosed(t *testing.T) {
	b1 := []byte("one for all\n")
	b2 := []byte("all for one\n")
	input1 := bytes.NewBuffer(b1)
	input2 := bytes.NewBuffer(b2)
	match := blush.NewExact("one", blush.Blue)
	r := ioutil.NopCloser(io.MultiReader(input1, input2))
	b := &blush.Blush{
		Finders: []blush.Finder{match},
		Reader:  r,
	}
	p := make([]byte, len(b1))
	_, err := b.Read(p)
	if err != nil {
		t.Error(err)
	}
	err = b.Close()
	if err != nil {
		t.Fatal(err)
	}
	n, err := b.Read(p)
	if n != 0 {
		t.Errorf("b.Read(): n = %d, want 0", n)
	}
	if err != blush.ErrClosed {
		t.Errorf("b.Read(): err = %v, want %v", err, blush.ErrClosed)
	}
}

func TestBlushReadLongOneLineText(t *testing.T) {
	head := strings.Repeat("a", 10000)
	tail := strings.Repeat("b", 10000)
	input := bytes.NewBuffer([]byte(head + " FINDME " + tail))
	match := blush.NewExact("FINDME", blush.Blue)
	r := ioutil.NopCloser(input)
	b := &blush.Blush{
		Finders: []blush.Finder{match},
		Reader:  r,
	}
	p := make([]byte, 20)
	_, err := b.Read(p)
	if err != nil {
		t.Error(err)
	}
	err = b.Close()
	if err != nil {
		t.Fatal(err)
	}
	n, err := b.Read(p)
	if n != 0 {
		t.Errorf("b.Read(): n = %d, want 0", n)
	}
	if err != blush.ErrClosed {
		t.Errorf("b.Read(): err = %v, want %v", err, blush.ErrClosed)
	}
}

func TestPrintName(t *testing.T) {
	line1 := "line one\n"
	line2 := "line two\n"
	r1 := ioutil.NopCloser(bytes.NewBuffer([]byte(line1)))
	r2 := ioutil.NopCloser(bytes.NewBuffer([]byte(line2)))
	name1 := "reader1"
	name2 := "reader2"
	r, err := blush.NewMultiReader(
		blush.WithReader(name1, r1),
		blush.WithReader(name2, r2),
	)
	if err != nil {
		t.Fatal(err)
	}
	b := blush.Blush{
		Reader:       r,
		Finders:      []blush.Finder{blush.NewExact("line", blush.NoColour)},
		WithFileName: true,
	}
	buf := new(bytes.Buffer)
	n, err := b.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}
	total := len(line1+line2+name1+name2) + len(blush.Separator)*2
	if int(n) != total {
		t.Errorf("total reads = %d, want %d", n, total)
	}
	s := strings.Split(buf.String(), "\n")
	if !strings.Contains(s[0], name1) {
		t.Errorf("want `%s` in `%s`", name1, s[0])
	}
	if !strings.Contains(s[1], name2) {
		t.Fatalf("want `%s` in `%s`", name2, s[1])
	}

}

func TestPrintFileName(t *testing.T) {
	path, err := ioutil.TempDir("", "blush_name")
	if err != nil {
		t.Fatal(err)
	}
	f1, err := ioutil.TempFile(path, "blush_name")
	if err != nil {
		t.Fatal(err)
	}
	f2, err := ioutil.TempFile(path, "blush_name")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(path); err != nil {
			t.Error(err)
		}
	}()
	line1 := "line one\n"
	line2 := "line two\n"
	f1.WriteString(line1)
	f2.WriteString(line2)
	tcs := []struct {
		name          string
		withFilename  bool
		wantLen       int
		wantFilenames bool
	}{
		{"with filename", true, len(line1+line2+f1.Name()+f2.Name()) + len(blush.Separator)*2, true},
		{"without filename", false, len(line1 + line2), false},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r, err := blush.NewMultiReader(
				blush.WithPaths([]string{path}, false),
			)
			if err != nil {
				t.Fatal(err)
			}
			b := blush.Blush{
				Reader:       r,
				Finders:      []blush.Finder{blush.NewExact("line", blush.NoColour)},
				WithFileName: tc.withFilename,
			}
			buf := new(bytes.Buffer)
			n, err := b.WriteTo(buf)
			if err != nil {
				t.Fatal(err)
			}
			if int(n) != tc.wantLen {
				t.Errorf("total reads = %d, want %d", n, tc.wantLen)
			}
			notStr := "not"
			if tc.wantFilenames {
				notStr = ""
			}
			if strings.Contains(buf.String(), f1.Name()) != tc.wantFilenames {
				t.Errorf("want `%s` %s in `%s`", f1.Name(), notStr, buf.String())
			}
			if strings.Contains(buf.String(), f2.Name()) != tc.wantFilenames {
				t.Errorf("want `%s` %s in `%s`", f2.Name(), notStr, buf.String())
			}
		})
	}
}
