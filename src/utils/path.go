package utils

import (
	"path/filepath"
	"runtime"
	"strings"
)

func CleanFilePathAbs(pathstr string) (string, error) {
	pathstr, err := filepath.Abs(filepath.Clean(pathstr))
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		index := strings.Index(pathstr, `:\`)
		pf := strings.ToUpper(pathstr[:index])
		ph := pathstr[index:]
		pathstr = pf + ph
	}

	return pathstr, nil
}

func FilePathEqual(path1, path2 string) bool {
	path1, err := CleanFilePathAbs(path1)
	if err != nil {
		return false
	}

	path2, err = CleanFilePathAbs(path2)
	if err != nil {
		return false
	}

	return path1 == path2
}

func CheckIfSubPath(parentPath, childPath string) bool {
	parentPath, err := CleanFilePathAbs(parentPath)
	if err != nil {
		return false
	}

	childPath, err = CleanFilePathAbs(childPath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(childPath, parentPath)
}

func CheckIfSubPathNotEqual(parentPath, childPath string) bool {
	parentPath, err := CleanFilePathAbs(parentPath)
	if err != nil {
		return false
	}

	childPath, err = CleanFilePathAbs(childPath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(childPath, parentPath) && childPath != parentPath
}
