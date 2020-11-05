package zipper

import (
	"archive/zip"
	"io/ioutil"
	"os"
)

// File 方便自定义每个文件的文件名。
type File struct {
	Name string
	Path string
}

// Create 创建一个压缩包，把 files 打包到 zipFilePath.
func Create(zipFilePath string, files []File) error {
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	for _, file := range files {
		addFile(file, zipWriter)
	}

	// 如果发生错误，要删除文件。
	if err := zipWriter.Close(); err != nil {
		zipFile.Close()
		if err := os.Remove(zipFilePath); err != nil {
			return err
		}
		return err
	}
	return nil
}

func addFile(file File, zipWriter *zip.Writer) error {
	fileIndeed, err := os.Open(file.Path)
	if err != nil {
		return err
	}
	defer fileIndeed.Close()

	body, err := ioutil.ReadAll(fileIndeed)
	if err != nil {
		return err
	}

	fileInZip, err := zipWriter.Create(file.Name)
	if err != nil {
		return err
	}

	if _, err = fileInZip.Write(body); err != nil {
		return err
	}
	return nil
}
