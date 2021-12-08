package jobs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/sayuri567/setu-checker/baiduaudit"
	"github.com/sayuri567/setu-checker/config"
	"github.com/sayuri567/tool/util/arrayutil"
	"github.com/sayuri567/tool/util/fileutil"
	"github.com/sirupsen/logrus"
)

type checkImages struct{}

var CheckImages = &checkImages{}

func (this *checkImages) Run() {
	client := baiduaudit.GetClient(config.Conf.BaseConf.BaiduAk, config.Conf.BaseConf.BaiduSk)
	for _, imagePath := range config.Conf.BaseConf.Paths {
		files, err := fileutil.GetAllFiles(imagePath)
		if err != nil {
			logrus.WithError(err).Error("failed to get all files")
			continue
		}
		for _, file := range files {
			tp := this.getFileType(file)
			if arrayutil.InArrayForString(config.Conf.BaseConf.FileType, strings.ToLower(tp)) == -1 {
				continue
			}
			resp, err := client.CheckImages(file.Path)
			if err != nil && !errors.Is(err, baiduaudit.ErrInvalidFileType) {
				logrus.WithError(err).Error("failed to check image")
				continue
			}
			err = this.classify(file, resp)
			if err != nil {
				logrus.WithError(err).Error("failed to classify image")
				continue
			}
		}
	}
}

func (this *checkImages) moveFile(file *fileutil.File, targetDir string) error {
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

	if targetDir != config.Conf.AuditConf.FailDir {
		return this.moveFile(file, config.Conf.AuditConf.FailDir)
	}
	return errors.New("move file fail!!!")
}

func (this *checkImages) classify(file *fileutil.File, resp *baiduaudit.CheckImageResp) error {
	// 不检测
	if resp == nil {
		return this.moveFile(file, config.Conf.AuditConf.NoCheck)
	}
	// 普通图片
	if resp.ConclusionType == 1 {
		return this.moveFile(file, config.Conf.AuditConf.NoH)
	}
	if resp.ConclusionType == 2 || resp.ConclusionType == 3 {
		sort.Sort(sort.Reverse(resp.Data))
		for _, item := range resp.Data {
			// 非涩情审核内容
			if item.Type != 1 {
				continue
			}
			dir := ""
			switch item.SubType {
			case 0: // 一般涩情
				dir = config.Conf.AuditConf.NormalH
			case 1: // 卡通涩情
				dir = config.Conf.AuditConf.AnimeH
			case 2: // SM
				dir = config.Conf.AuditConf.SM
			case 3: // 低俗
				dir = config.Conf.AuditConf.LowH
			case 4: // LOLI
				dir = config.Conf.AuditConf.LoliH
			case 5: // 艺术品涩情
				dir = config.Conf.AuditConf.ArtH
			case 6: // 性玩具
				dir = config.Conf.AuditConf.ToysH
			case 7: // 男性性感
				dir = config.Conf.AuditConf.MenSexyH
			case 8: // 男性裸露
				dir = config.Conf.AuditConf.MenBareH
			case 9: // 女性性感
				dir = config.Conf.AuditConf.NormalSexyH
			case 10: // 卡通女性性感
				dir = config.Conf.AuditConf.AnimeSexyH
			case 11: // 孕妇
				dir = config.Conf.AuditConf.PregnantH
			case 12: // 特殊
				dir = config.Conf.AuditConf.SpecialH
			case 13: // 臀部特写
				dir = config.Conf.AuditConf.HipsH
			case 14: // 脚部特写
				dir = config.Conf.AuditConf.FeetH
			case 15: // 裆部特写
				dir = config.Conf.AuditConf.CrotchH
			case 16: // 亲密行为
				dir = config.Conf.AuditConf.IntimateH
			case 17: // 卡通亲密行为
				dir = config.Conf.AuditConf.AnimeIntimateH
			default: // 其他类型
				dir = config.Conf.AuditConf.FailDir
			}

			return this.moveFile(file, path.Join(dir, this.getScore(item.Probability)))
		}
	}

	return this.moveFile(file, config.Conf.AuditConf.FailDir)
}

func (this *checkImages) getScore(probability float64) string {
	p := probability * 100
	if p >= 0 && p < 20 {
		return "20"
	} else if p >= 20 && p < 40 {
		return "40"
	} else if p >= 40 && p < 60 {
		return "60"
	} else if p >= 60 && p < 80 {
		return "80"
	} else if p >= 80 && p <= 100 {
		return "100"
	}
	return "0"
}

func (this *checkImages) getFileType(file *fileutil.File) string {
	idx := strings.LastIndex(file.Path, ".")
	if idx == -1 || idx+1 >= len(file.Path) {
		return ""
	}
	return file.Path[idx+1:]
}