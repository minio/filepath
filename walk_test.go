// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the https://golang.org/LICENSE file.

package filepath

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestWalkIsFlatKeySorted checks the order this package exists to produce:
//
//	lexically ordered output treats directories with `/` at the end before
//	sorting, this is sorting is flat key sorting instead of regular
//	filesystem sorting of entries.
//
// Flat key order is an object store's order. A key `a/b` sorts under `a`
// because `/` is 0x2F, below every character a sibling name can start with.
// byName appends os.PathSeparator rather than `/`: the same byte on Unix, and
// 0x5C on Windows, which sits above the digits and the upper case letters.
//
// So one directory and its siblings are ordered differently on Windows, and
// the walk there is neither flat key sorted nor in the order an object store
// lists the same names in. The tree below puts a directory among siblings
// whose first differing byte falls on each side of the two separators:
//
//	0x2D  a-1.bin   below `/` and below `\`
//	0x2F  a/        the directory, where flat key order puts it
//	0x30  a0.bin    above `/`, below `\`  -- moves on Windows
//	0x5A  aZ.bin    above `/`, below `\`  -- moves on Windows
//	0x5C  a\        the directory, where this walk puts it on Windows
//	0x5F  a_.bin    above both
//	0x61  aa.bin    above both
//
// Every name is reported in slash form, so what is compared here is the order
// alone and not how a platform spells a path.
func TestWalkIsFlatKeySorted(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"a/inner.bin",
		"a-1.bin",
		"a0.bin",
		"aZ.bin",
		"a_.bin",
		"aa.bin",
	} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var got []string
	err := Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		if rel == "." {
			return nil
		}
		got = append(got, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"a-1.bin", "a", "a/inner.bin", "a0.bin", "aZ.bin", "a_.bin", "aa.bin"}

	t.Logf("GOOS=%s, os.PathSeparator=%q (0x%02X)",
		runtime.GOOS, os.PathSeparator, os.PathSeparator)
	t.Logf("flat key order: %s", strings.Join(want, "  "))
	t.Logf("this walk     : %s", strings.Join(got, "  "))

	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("the walk is not flat key sorted on %s\n"+
			"  want: %s\n"+
			"  got:  %s\n"+
			"byName appends os.PathSeparator (%q) where the package comment says `/`, "+
			"so the directory sorts above every sibling whose next byte is below 0x%02X.",
			runtime.GOOS, strings.Join(want, "  "), strings.Join(got, "  "),
			os.PathSeparator, os.PathSeparator)
	}
}
