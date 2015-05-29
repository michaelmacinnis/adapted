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

func findExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}

func LookPath(file string) (string, error) {
	cnf := "command not found"

	// Only bypass the path if file begins with / or ./ or ../
	prefix := file + "   "
	if prefix[0:1] == "/" || prefix[0:2] == "./" || prefix[0:3] == "../" {
		err := findExecutable(file)
		if err == nil {
			return file, nil
		}
		return "", &pathError{file, err.Error()}
	}
	pathenv := os.Getenv("PATH")
	if pathenv == "" {
		return "", &pathError{file, cnf}
	}
	for _, dir := range strings.Split(pathenv, ":") {
		path := dir + "/" + file
		if err := findExecutable(path); err == nil {
			return path, nil
		}
	}
	return "", &pathError{file, cnf}
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
