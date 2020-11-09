package common // import "github.com/ahui2016/go-send/common"

import "path/filepath"

const (
	DataFolderName   = "gosend_data_folder"
	FilesFolderName  = "files"
	DatabaseFileName = "gosend.db"
	GosendFileExt    = ".send"
	ThumbFileExt     = ".small"

	// 99 days, for session
	MaxAge = 60 * 60 * 24 * 99

	// DatabaseCapacity 控制数据库总容量，
	// maxBodySize 控制单个文件的体积。
	DatabaseCapacity = 1 << 30 // 1GB
)

// LocalFilePath .
func LocalFilePath(filesDir, id string) string {
	return filepath.Join(filesDir, id+GosendFileExt)
}

// ThumbFilePath .
func ThumbFilePath(filesDir, id string) string {
	return filepath.Join(filesDir, id+ThumbFileExt)
}

// GetFileAndThumb .
func GetFileAndThumb(filesDir, id string) (originFile, thumb string) {
	return LocalFilePath(filesDir, id), ThumbFilePath(filesDir, id)
}
