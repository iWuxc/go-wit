package utils

import (
	"fmt"
	"github.com/iWuxc/go-wit/errors"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func FileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetPath(path string) string {
	return filepath.Dir(path)
}

// OpenFileOrStdDev .
func OpenFileOrStdDev(path string, write bool) (*os.File, error) {
	var fp *os.File
	var err error

	switch path {
	case "stdin", "STDIN":
		fp = os.Stdin
	case "stdout", "STDOUT":
		fp = os.Stdout
	default:
		if write {
			fp, err = os.Create(CleanPath(path))
		} else {
			fp, err = os.Open(CleanPath(path))
		}
	}

	if err != nil {
		return nil, err
	}

	stat, _ := fp.Stat()
	if stat.IsDir() {
		return nil, errors.Errorf("%s: is a directory", path)
	}
	return fp, nil
}

// DirectoryFiles get files in directory .
func DirectoryFiles(path string) ([]string, error) {
	cleanPath := CleanPath(path)
	directoryEntries, err := ioutil.ReadDir(cleanPath)
	result := make([]string, 0)
	if err != nil {
		return nil, err
	}

	for _, entry := range directoryEntries {
		if !entry.IsDir() {
			result = append(result, filepath.Join(cleanPath, entry.Name()))
		}
	}

	return result, nil
}

// IsDir .
func IsDir(path string) (bool, error) {
	fp, err := os.Open(CleanPath(path))
	if err != nil {
		return false, err
	}
	defer fp.Close()

	stat, statErr := fp.Stat()
	if statErr != nil {
		return false, statErr
	}
	return stat.IsDir(), nil
}

func CleanPath(path string) string {
	if path == "" {
		return ""
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	result := ""
	if len(path) > 1 && path[:2] == "~/" {
		dir := usr.HomeDir + "/"
		result = strings.Replace(path, "~/", dir, 1)
	} else {
		result = path
	}

	absResult, _ := filepath.Abs(result)
	cleanResult := filepath.Clean(absResult)
	return cleanResult
}

func EnsureFileFolderExists(path string) error {
	p := GetPath(path)
	if !FileExist(p) {
		err := os.MkdirAll(p, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveExt(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

func UrlJoin(base string, path string) string {
	res := fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(path, "/"))
	return res
}

func GetUrlPath(urlString string) string {
	u, _ := url.Parse(urlString)
	return u.Path
}

func GetUrlHost(urlString string) string {
	u, _ := url.Parse(urlString)
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func FilterQuery(urlString string, blackList []string) string {
	urlData, err := url.Parse(urlString)
	if err != nil {
		return urlString
	}

	queries := urlData.Query()
	retQuery := make(url.Values)
	inBlackList := false
	for key, value := range queries {
		inBlackList = false
		for _, blackListItem := range blackList {
			if blackListItem == key {
				inBlackList = true
				break
			}
		}
		if !inBlackList {
			retQuery[key] = value
		}
	}
	if len(retQuery) > 0 {
		return urlData.Path + "?" + strings.ReplaceAll(retQuery.Encode(), "%2F", "/")
	} else {
		return urlData.Path
	}
}
