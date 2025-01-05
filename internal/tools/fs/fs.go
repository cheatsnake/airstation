package fs

import "os"

func MustNewDir(dirPath string) {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		panic(err)
	}
}
