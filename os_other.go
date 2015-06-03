// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !linux,!darwin,!dragonfly,!freebsd,!openbsd,!netbsd,!solaris

package adapted

import (
	"fmt"
)

func TempFifo(prefix string) (name string, err error) {
	return "", fmt.Errorf("TempFifo not implemented")
}

