package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sayuri567/gorun"
	"github.com/sayuri567/setu-checker/baiduaudit"
	"github.com/sayuri567/setu-checker/config"
	"github.com/sayuri567/tool/util/arrayutil"
	"github.com/sayuri567/tool/util/fileutil"
	"github.com/sirupsen/logrus"
)

type checkImages struct {
	baiduClient *baiduaudit.Client
	ignoreRegs  []*regexp.Regexp
	workers     *gorun.GoRun
}

type statisticsData struct {
	FileCount    int64    `json:"file_count"`
	IgnoreCount  int64    `json:"ignore_count"`
	VideoCount   int64    `json:"video_count"`
	GifCount     int64    `json:"gif_count"`
	ImageCount   int64    `json:"image_count"`
	FailCount    int64    `json:"fail_count"`
	InvalidCount int64    `json:"invalid_count"`
	InvalidTypes []string `json:"invalid_types"`
}

var CheckImages = &checkImages{}

func (this *checkImages) Run() {
	err := this.init()
	if err != nil {
		logrus.WithError(err).Error("failed to init check images")
		return
	}
	wg := &sync.WaitGroup{}

	statisticsMap := map[string]*statisticsData{}
	for _, imagePath := range config.Conf.BaseConf.Paths {
		files, err := fileutil.GetAllFiles(imagePath, config.Conf.BaseConf.IgnoreDir...)
		if err != nil {
			logrus.WithError(err).Error("failed to get all files")
			continue
		}
		wg.Add(len(files))
		statisticsMap[imagePath] = &statisticsData{FileCount: int64(len(files))}
		for _, file := range files {
			this.workers.Go(this.doFileClassify, file, wg, statisticsMap[imagePath])
		}
	}
	wg.Wait()

	for _, item := range statisticsMap {
		item.InvalidTypes = arrayutil.UniqueString(item.InvalidTypes)
	}

	logrus.WithField("statistics", statisticsMap).Info("check images statistics")
}

// 初始化
func (this *checkImages) init() error {
	if this.baiduClient == nil {
		this.baiduClient = baiduaudit.GetClient(config.Conf.BaseConf.BaiduAk, config.Conf.BaseConf.BaiduSk)
	}
	if len(this.ignoreRegs) == 0 && len(config.Conf.BaseConf.IgnoreFile) > 0 {
		this.ignoreRegs = []*regexp.Regexp{}
		for _, ignoreFile := range config.Conf.BaseConf.IgnoreFile {
			this.ignoreRegs = append(this.ignoreRegs, regexp.MustCompile(ignoreFile))
		}
	}
	if this.workers == nil {
		worker := 2
		if config.Conf.BaseConf.Worker > 0 {
			worker = config.Conf.BaseConf.Worker
		}
		this.workers = gorun.NewGoRun(worker, time.Minute*10, true)
	}

	return nil
}

// 移动文件
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

	return errors.New("move file fail!!!")
}

func (this *checkImages) doFileClassify(file *fileutil.File, wg *sync.WaitGroup, statistics *statisticsData) {
	defer wg.Done()
	isIgnore := false
	var err error
	for _, reg := range this.ignoreRegs {
		isIgnore = reg.MatchString(file.Name)
		if isIgnore {
			break
		}
	}
	if isIgnore {
		return
	}
	tp := this.getFileType(file)
	switch true {
	case arrayutil.InArrayForString(config.Conf.BaseConf.GifType, strings.ToLower(tp)) > -1:
		atomic.AddInt64(&statistics.GifCount, 1)
		err = this.classifyGif(file)
	case arrayutil.InArrayForString(config.Conf.BaseConf.VideoType, strings.ToLower(tp)) > -1:
		atomic.AddInt64(&statistics.VideoCount, 1)
		err = this.classifyVideo(file)
	case arrayutil.InArrayForString(config.Conf.BaseConf.FileType, strings.ToLower(tp)) > -1:
		atomic.AddInt64(&statistics.ImageCount, 1)
		err = this.classifyImg(file)
	default:
		atomic.AddInt64(&statistics.InvalidCount, 1)
		statistics.InvalidTypes = append(statistics.InvalidTypes, strings.ToLower(tp))
		// err = this.classifyDefault(file)
	}
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"file_type": strings.ToLower(tp), "file_path": file.Path}).Error("failed to classify image")
		atomic.AddInt64(&statistics.FailCount, 1)
		this.classifyFail(file)
		return
	}
}

func (this *checkImages) classifyGif(file *fileutil.File) error {
	return this.moveFile(file, config.Conf.AuditConf.Gif)
}

func (this *checkImages) classifyVideo(file *fileutil.File) error {
	return this.moveFile(file, config.Conf.AuditConf.Mp4)
}

func (this *checkImages) classifyImg(file *fileutil.File) error {
	resp, err := this.baiduClient.CheckImages(file.Path)
	if err != nil {
		logrus.WithError(err).Error("failed to check image")
		return err
	}
	if resp.ErrorCode > 0 {
		logrus.WithField("resp", resp).Error("check image resp has error")
		return errors.New(resp.ErrorMsg)
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
			case 11: // 特殊
				dir = config.Conf.AuditConf.SpecialH
			case 12: // 亲密行为
				dir = config.Conf.AuditConf.IntimateH
			case 13: // 卡通亲密行为
				dir = config.Conf.AuditConf.AnimeIntimateH
			case 14: // 孕妇
				dir = config.Conf.AuditConf.PregnantH
			case 15: // 臀部特写
				dir = config.Conf.AuditConf.HipsH
			case 16: // 脚部特写
				dir = config.Conf.AuditConf.FeetH
			case 17: // 裆部特写
				dir = config.Conf.AuditConf.CrotchH
			default: // 其他类型
				dir = config.Conf.AuditConf.Other
			}

			return this.moveFile(file, path.Join(dir, this.getScore(item.Probability)))
		}
	}

	errMsg, _ := json.Marshal(resp)
	return fmt.Errorf("audit fail, %s", string(errMsg))
}

func (this *checkImages) classifyDefault(file *fileutil.File) error {
	return this.moveFile(file, config.Conf.AuditConf.NoCheck)
}

func (this *checkImages) classifyFail(file *fileutil.File) error {
	return this.moveFile(file, config.Conf.AuditConf.FailDir)
}

// 获取百度接口分数
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

// 获取文件类型
func (this *checkImages) getFileType(file *fileutil.File) string {
	idx := strings.LastIndex(file.Path, ".")
	if idx == -1 || idx+1 >= len(file.Path) {
		return ""
	}
	return file.Path[idx+1:]
}
