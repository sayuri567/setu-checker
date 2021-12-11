package filesource

import (
	"sync"

	"github.com/sayuri567/tool/util/fileutil"
)

var sources = map[string]FileSource{}
var sourceLock = &sync.RWMutex{}

type FileSource interface {
	GetAllFiles(dirPath string, ignoreDirName ...string) (files fileutil.Files, err error)
	MoveFile(file *fileutil.File, targetDir string) error
	GetFile(file *fileutil.File) ([]byte, error)
}

func RegisterSource(tp string, source FileSource) {
	sourceLock.Lock()
	defer sourceLock.Unlock()
	sources[tp] = source
}

func GetSource(tp string) FileSource {
	sourceLock.RLock()
	defer sourceLock.RUnlock()
	return sources[tp]
}
