package fs

import (
	"fmt"
	"os"
)

func MustDir(dirPath string) {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		panic(err)
	}
}

func FileExists(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %v", filePath)
	}

	if err != nil {
		return fmt.Errorf("retrieving file info failed: %s", err.Error())
	}

	return nil
}

func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("deleting file failed: %s", err.Error())
	}

	return nil
}

func DeleteFolderIfExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking folder: %v", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to delete folder: %v", err)
	}

	return nil
}

func RenameFile(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("renaming file failed: %s", err.Error())
	}

	return nil
}
