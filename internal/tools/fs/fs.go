package fs

import "os"

func MustDir(dirPath string) {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		panic(err)
	}
}
