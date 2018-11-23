package utils

import (
	"os"
)

func CreateIfNotExist(dirName string) error {
	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		return os.Mkdir("logFiles", os.ModePerm)
	}

	return err
}
