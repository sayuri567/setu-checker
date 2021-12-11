package filesource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/sayuri567/setu-checker/config"
	"github.com/sayuri567/tool/util/arrayutil"
	"github.com/sayuri567/tool/util/fileutil"
)

const SOURCE_FTP = "FTP"

type FTPSource struct {
	conn *ftp.ServerConn
}

func init() {
	RegisterSource(SOURCE_FTP, &FTPSource{})
}

func (this *FTPSource) GetAllFiles(dirPath string, ignoreDirName ...string) (fileutil.Files, error) {
	if err := this.checkConn(); err != nil {
		return nil, err
	}
	entrys, err := this.conn.List(dirPath)
	if err != nil {
		return nil, err
	}

	files := fileutil.Files{}
	for _, entry := range entrys {
		if entry.Type == ftp.EntryTypeFile {
			files = append(files, &fileutil.File{
				Name:    entry.Name,
				Path:    path.Join(dirPath, entry.Name),
				ModTime: entry.Time,
				Size:    int64(entry.Size),
			})
		} else if entry.Type == ftp.EntryTypeFolder {
			if arrayutil.InArrayForString(ignoreDirName, entry.Name) > -1 {
				continue
			}
			childFiles, err := this.GetAllFiles(path.Join(dirPath, entry.Name), ignoreDirName...)
			if err != nil {
				return nil, err
			}
			files = append(files, childFiles...)
		}
	}

	return files, nil
}

func (this *FTPSource) GetFile(file *fileutil.File) ([]byte, error) {
	resp, err := this.conn.Retr(file.Path)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	return ioutil.ReadAll(resp)
}

func (this *FTPSource) MoveFile(file *fileutil.File, targetDir string) error {
	if err := this.checkConn(); err != nil {
		return err
	}

	if len(targetDir) == 0 {
		return errors.New("Empty targetDir")
	}

	err := this.conn.ChangeDir(targetDir)
	if err != nil {
		t, _ := this.conn.GetTime(targetDir)
		if !t.IsZero() {
			return fmt.Errorf("targetDir is not dir: %s", targetDir)
		}
		err = this.conn.MakeDir(targetDir)
		if err != nil {
			return err
		}
	}
	this.conn.ChangeDir("/")

	newName := path.Join(targetDir, file.Name)

	for i := 0; i < 3; i++ {
		size, _ := this.conn.FileSize(newName)
		if size > 0 {
			idx := strings.LastIndex(newName, ".")
			newName = fmt.Sprintf("%s_%v%s", newName[:idx], time.Now().Unix(), newName[idx:])
			continue
		}
		break
	}
	return this.conn.Rename(file.Path, newName)
}

func (this *FTPSource) checkConn() error {
	var err error
	if this.conn != nil {
		err = this.conn.NoOp()
		if err == nil {
			return nil
		}
	}

	this.conn, err = ftp.Dial(config.Conf.FtpConf.Address, ftp.DialWithTimeout(time.Second*5))
	if err != nil {
		return err
	}

	return this.conn.Login(config.Conf.FtpConf.Username, config.Conf.FtpConf.Password)
}
