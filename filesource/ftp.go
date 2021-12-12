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
const maxConn = 5

type FTPSource struct {
	conns chan *ftpConn
}

type ftpConn struct {
	conn   *ftp.ServerConn
	source *FTPSource
}

func init() {
	source := &FTPSource{conns: make(chan *ftpConn, maxConn)}
	for i := 0; i < maxConn; i++ {
		source.conns <- &ftpConn{}
	}
	RegisterSource(SOURCE_FTP, source)
}

func (this *FTPSource) GetAllFiles(dirPath string, ignoreDirName ...string) (fileutil.Files, error) {
	conn, err := this.getConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	entrys, err := conn.conn.List(dirPath)
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
	conn, err := this.getConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	resp, err := conn.conn.Retr(file.Path)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	return ioutil.ReadAll(resp)
}

func (this *FTPSource) MoveFile(file *fileutil.File, targetDir string) error {
	conn, err := this.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	if len(targetDir) == 0 {
		return errors.New("Empty targetDir")
	}

	err = conn.conn.ChangeDir(targetDir)
	if err != nil {
		t, _ := conn.conn.GetTime(targetDir)
		if !t.IsZero() {
			return fmt.Errorf("targetDir is not dir: %s", targetDir)
		}
		err = conn.conn.MakeDir(targetDir)
		if err != nil {
			return err
		}
	}
	conn.conn.ChangeDir("/")

	newName := path.Join(targetDir, file.Name)

	for i := 0; i < 3; i++ {
		size, _ := conn.conn.FileSize(newName)
		if size > 0 {
			idx := strings.LastIndex(newName, ".")
			newName = fmt.Sprintf("%s_%v%s", newName[:idx], time.Now().Unix(), newName[idx:])
			continue
		}
		break
	}
	return conn.conn.Rename(file.Path, newName)
}

func (this *FTPSource) getConn() (*ftpConn, error) {
	conn := <-this.conns
	var err error
	if conn.conn == nil {
		conn, err = this.createConn()
	}
	if err = conn.conn.NoOp(); err != nil {
		conn.conn.Quit()
		conn, err = this.createConn()
	}
	if err != nil {
		this.conns <- conn
		return nil, err
	}

	return conn, nil
}

func (this *FTPSource) createConn() (*ftpConn, error) {
	conn, err := ftp.Dial(config.Conf.FtpConf.Address)
	if err != nil {
		return nil, err
	}
	err = conn.Login(config.Conf.FtpConf.Username, config.Conf.FtpConf.Password)
	if err != nil {
		return nil, err
	}

	return &ftpConn{conn: conn, source: this}, nil
}

func (this *ftpConn) Close() {
	this.source.conns <- this
}
