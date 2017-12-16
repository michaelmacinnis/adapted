// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package adapted

import (
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

type pathError struct {
	Path string
	Err  string
}

func (e *pathError) Error() string {
	return e.Path + ": " + e.Err
}

const lowerhex = "0123456789abcdef"

func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

func findPath(file string) (bool, error) {
	d, err := os.Stat(file)
	if err != nil {
		return false, err
	}

	m := d.Mode()
	if m.IsDir() {
		return false, nil
	} else if m&0111 != 0 {
		return true, nil
	}
	return false, os.ErrPermission
}

func LookPath(pathenv, file string) (string, bool, error) {
	cnf := "command not found"

	// Only bypass the path if file begins with / or ./ or ../
	prefix := file + "   "
	if prefix[0:1] == "/" || prefix[0:2] == "./" || prefix[0:3] == "../" {
		exe, err := findPath(file)
		if err == nil {
			return file, exe, nil
		}
		return "", false, &pathError{file, err.Error()}
	}
	if pathenv == "" {
		return "", false, &pathError{file, cnf}
	}
	for _, dir := range strings.Split(pathenv, ":") {
		path := dir + "/" + file
		if exe, err := findPath(path); err == nil {
			return path, exe, nil
		}
	}
	return "", false, &pathError{file, cnf}
}

func Quote(s string) string {
	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	buf = append(buf, '"')
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		if r == rune('"') || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		case '\a', '\b', '\f', '\n', '\r', '\t', '\v':
			buf = append(buf, byte(r))
		default:
			switch {
			case r < ' ':
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[0]>>4])
				buf = append(buf, lowerhex[s[0]&0xF])
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				buf = append(buf, `\u`...)
				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				buf = append(buf, `\U`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}
	buf = append(buf, '"')
	return string(buf)
}

func Unquote(s string) (t string, err error) {
	n := len(s)
	if n < 2 {
		return "", strconv.ErrSyntax
	}
	quote := s[0]
	if quote != s[n-1] {
		return "", strconv.ErrSyntax
	}
	s = s[1 : n-1]

	if quote != '"' {
		return "", strconv.ErrSyntax
	}

	// Is it trivial?  Avoid allocation.
	if !contains(s, '\\') && !contains(s, quote) {
		switch quote {
		case '"':
			return s, nil
		case '\'':
			r, size := utf8.DecodeRuneInString(s)
			if size == len(s) && (r != utf8.RuneError || size != 1) {
				return s, nil
			}
		}
	}

	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		c, multibyte, ss, err := strconv.UnquoteChar(s, quote)
		if err != nil {
			s = s[1:]
			continue
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
	}
	return string(buf), nil
}

