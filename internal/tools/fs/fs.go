package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func DeleteDirIfExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking directory: %v", err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to delete directory: %v", err)
	}

	return nil
}

func RenameFile(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("renaming file failed: %v", err.Error())
	}

	return nil
}

func ListFilesFromDir(dirPath, fileExt string) ([]string, error) {
	var filenames []string

	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %v", err)
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if strings.Contains(ext, fileExt) {
				relPath, err := filepath.Rel(dirPath, path)
				if err != nil {
					return err
				}
				filenames = append(filenames, relPath)
			}
		}
		return nil
	})

	return filenames, err
}
