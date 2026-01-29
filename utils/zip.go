package utils

import (
	"archive/zip"
	"io"
	"os"
)

// Unzip 解压压缩包文件.
func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fPath := AppendPath(destDir, f.Name)
		_ = EnsureFileFolderExists(fPath)
		if !f.FileInfo().IsDir() {
			inFile, err := f.Open()
			if err != nil {
				return err
			}

			outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			if _, err = io.Copy(outFile, inFile); err != nil {
				return err
			}
			inFile.Close()
			outFile.Close()
		}
	}
	return nil
}
