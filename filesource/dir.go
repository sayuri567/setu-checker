package filesource

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/sayuri567/tool/util/fileutil"
)

const SOURCE_DIR = "DIR"

type DirSource struct{}

func init() {
	RegisterSource(SOURCE_DIR, &DirSource{})
}

func (this *DirSource) GetAllFiles(dirPath string, ignoreDirName ...string) (fileutil.Files, error) {
	return fileutil.GetAllFiles(dirPath, ignoreDirName...)
}

func (this *DirSource) GetFile(file *fileutil.File) ([]byte, error) {
	return os.ReadFile(file.Name)
}

func (this *DirSource) MoveFile(file *fileutil.File, targetDir string) error {
	if len(targetDir) == 0 {
		return errors.New("Empty targetDir")
	}
	dirInfo, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = fileutil.MakeDir(targetDir)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if dirInfo != nil && !dirInfo.IsDir() {
		return fmt.Errorf("targetDir is not dir: %s", targetDir)
	}
	newName := path.Join(targetDir, file.Name)

	for i := 0; i < 3; i++ {
		if _, err := os.Stat(newName); err != nil && os.IsNotExist(err) {
			cmd := exec.Command("mv", file.Path, newName)
			_, err = cmd.Output()
			return err
		}
		idx := strings.LastIndex(newName, ".")
		newName = fmt.Sprintf("%s_%v%s", newName[:idx], time.Now().Unix(), newName[idx:])
	}

	return errors.New("move file fail!!!")
}
