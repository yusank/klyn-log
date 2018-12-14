// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package utils

import (
	"os"
)

// CreateIfNotExist -  check dir and created if not exist
func CreateIfNotExist(dirName string) error {
	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		return os.Mkdir("logFiles", os.ModePerm)
	}

	return err
}
