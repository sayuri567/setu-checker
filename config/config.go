package config

import (
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BaseConf  BaseConf  `yaml:"base_conf"`
	AuditConf AuditConf `yaml:"audit_conf"`
}

type BaseConf struct {
	Worker     int      `yaml:"worker"`
	Paths      []string `yaml:"paths"`
	BaiduAk    string   `yaml:"baidu_ak"`
	BaiduSk    string   `yaml:"baidu_sk"`
	FileType   []string `yaml:"file_type"`
	IgnoreFile []string `yaml:"ignore_file"`
	VideoType  []string `yaml:"video_type"`
	GifType    []string `yaml:"gif_type"`
}

type AuditConf struct {
	FailDir        string `yaml:"fail_dir"`         // 移动文件失败目录
	Mp4            string `yaml:"mp4"`              // 通过审核     subType:
	Gif            string `yaml:"gif"`              // 通过审核     subType:
	NoCheck        string `yaml:"no_check"`         // 没有检测
	NoH            string `yaml:"no_h"`             // 通过审核     subType:
	NormalH        string `yaml:"normal_h"`         // 一般涩情     subType: 0
	AnimeH         string `yaml:"anime_h"`          // 卡通涩情		subType: 1
	SM             string `yaml:"sm_h"`             // SM		   subType: 2
	LowH           string `yaml:"low_h"`            // 低俗		    subType: 3
	LoliH          string `yaml:"loli_h"`           // LOLI涩情		subType: 4
	ArtH           string `yaml:"art_h"`            // 艺术品涩情    subType: 5
	ToysH          string `yaml:"toys_h"`           // 性玩具		subType: 6
	MenSexyH       string `yaml:"men_sexy_h"`       // 男性性感		subType: 7
	MenBareH       string `yaml:"men_bare_h"`       // 男性裸露		subType: 8
	NormalSexyH    string `yaml:"normal_sexy_h"`    // 女性性感		subType: 9
	AnimeSexyH     string `yaml:"anime_sexy_h"`     // 卡通女性性感	 subType: 10
	SpecialH       string `yaml:"special_h"`        // 特殊类		subType: 11
	IntimateH      string `yaml:"intimate_h"`       // 亲密行为		subType: 12
	AnimeIntimateH string `yaml:"anime_intimate_h"` // 卡通亲密行为	 subType: 13
	PregnantH      string `yaml:"pregnant_h"`       // 孕妇			subType: 14
	HipsH          string `yaml:"hips_h"`           // 臀部特写		subType: 15
	FeetH          string `yaml:"feet_h"`           // 脚部特写		subType: 16
	CrotchH        string `yaml:"crotch_h"`         // 裆部特写		subType: 17
}

const (
	SETU_PATHS     = "SETU_PATHS"
	SETU_AK        = "SETU_AK"
	SETU_SK        = "SETU_SK"
	SETU_FILE_TYPE = "SETU_FILE_TYPE"

	SETU_FAIL_PATH             = "SETU_FAIL_PATH"
	SETU_NO_CHECK_PATH         = "SETU_NO_CHECK_PATH"
	SETU_NO_H_PATH             = "SETU_NO_H_PATH"
	SETU_NORMAL_H_PATH         = "SETU_NORMAL_H_PATH"
	SETU_ANIME_H_PATH          = "SETU_ANIME_H_PATH"
	SETU_SM_H_PATH             = "SETU_SM_H_PATH"
	SETU_LOW_H_PATH            = "SETU_LOW_H_PATH"
	SETU_LOLI_H_PATH           = "SETU_LOLI_H_PATH"
	SETU_ART_H_PATH            = "SETU_ART_H_PATH"
	SETU_TOYS_H_PATH           = "SETU_TOYS_H_PATH"
	SETU_MEN_SEXY_H_PATH       = "SETU_MEN_SEXY_H_PATH"
	SETU_MEN_BARE_H_PATH       = "SETU_MEN_BARE_H_PATH"
	SETU_NORMAL_SEXY_H_PATH    = "SETU_NORMAL_SEXY_H_PATH"
	SETU_ANIME_SEXY_H_PATH     = "SETU_ANIME_SEXY_H_PATH"
	SETU_PREGNANT_H_PATH       = "SETU_PREGNANT_H_PATH"
	SETU_SPECIAL_H_PATH        = "SETU_SPECIAL_H_PATH"
	SETU_HIPS_H_PATH           = "SETU_HIPS_H_PATH"
	SETU_FEET_H_PATH           = "SETU_FEET_H_PATH"
	SETU_CROTCH_H_PATH         = "SETU_CROTCH_H_PATH"
	SETU_INTIMATE_H_PATH       = "SETU_INTIMATE_H_PATH"
	SETU_ANIME_INTIMATE_H_PATH = "SETU_ANIME_INTIMATE_H_PATH"
)

var Conf = &Config{}

func (m *Config) Load(file string) error {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return m.loadEnv()
	}
	err = yaml.Unmarshal(yamlFile, m)
	if err != nil {
		return err
	}

	return m.loadEnv()
}

func (m *Config) loadEnv() error {
	paths := os.Getenv(SETU_PATHS)
	if len(paths) > 0 && len(m.BaseConf.Paths) == 0 {
		m.BaseConf.Paths = strings.Split(paths, ";")
	}
	ak := os.Getenv(SETU_AK)
	if len(ak) > 0 && len(m.BaseConf.BaiduAk) == 0 {
		m.BaseConf.BaiduAk = ak
	}
	sk := os.Getenv(SETU_SK)
	if len(sk) > 0 && len(m.BaseConf.BaiduSk) == 0 {
		m.BaseConf.BaiduSk = sk
	}
	tp := os.Getenv(SETU_FILE_TYPE)
	if len(tp) > 0 && len(m.BaseConf.FileType) == 0 {
		m.BaseConf.FileType = strings.Split(tp, ";")
	}

	path := os.Getenv(SETU_FAIL_PATH)
	if len(path) > 0 && len(m.AuditConf.FailDir) == 0 {
		m.AuditConf.FailDir = path
	}
	path = os.Getenv(SETU_NO_CHECK_PATH)
	if len(path) > 0 && len(m.AuditConf.NoCheck) == 0 {
		m.AuditConf.NoCheck = path
	}
	path = os.Getenv(SETU_NO_H_PATH)
	if len(path) > 0 && len(m.AuditConf.NoH) == 0 {
		m.AuditConf.NoH = path
	}
	path = os.Getenv(SETU_NORMAL_H_PATH)
	if len(path) > 0 && len(m.AuditConf.NormalH) == 0 {
		m.AuditConf.NormalH = path
	}
	path = os.Getenv(SETU_ANIME_H_PATH)
	if len(path) > 0 && len(m.AuditConf.AnimeH) == 0 {
		m.AuditConf.AnimeH = path
	}
	path = os.Getenv(SETU_SM_H_PATH)
	if len(path) > 0 && len(m.AuditConf.SM) == 0 {
		m.AuditConf.SM = path
	}
	path = os.Getenv(SETU_LOW_H_PATH)
	if len(path) > 0 && len(m.AuditConf.LowH) == 0 {
		m.AuditConf.LowH = path
	}
	path = os.Getenv(SETU_LOLI_H_PATH)
	if len(path) > 0 && len(m.AuditConf.LoliH) == 0 {
		m.AuditConf.LoliH = path
	}
	path = os.Getenv(SETU_ART_H_PATH)
	if len(path) > 0 && len(m.AuditConf.ArtH) == 0 {
		m.AuditConf.ArtH = path
	}
	path = os.Getenv(SETU_TOYS_H_PATH)
	if len(path) > 0 && len(m.AuditConf.ToysH) == 0 {
		m.AuditConf.ToysH = path
	}
	path = os.Getenv(SETU_MEN_SEXY_H_PATH)
	if len(path) > 0 && len(m.AuditConf.MenSexyH) == 0 {
		m.AuditConf.MenSexyH = path
	}
	path = os.Getenv(SETU_MEN_BARE_H_PATH)
	if len(path) > 0 && len(m.AuditConf.MenBareH) == 0 {
		m.AuditConf.MenBareH = path
	}
	path = os.Getenv(SETU_NORMAL_SEXY_H_PATH)
	if len(path) > 0 && len(m.AuditConf.NormalSexyH) == 0 {
		m.AuditConf.NormalSexyH = path
	}
	path = os.Getenv(SETU_ANIME_SEXY_H_PATH)
	if len(path) > 0 && len(m.AuditConf.AnimeSexyH) == 0 {
		m.AuditConf.AnimeSexyH = path
	}
	path = os.Getenv(SETU_PREGNANT_H_PATH)
	if len(path) > 0 && len(m.AuditConf.PregnantH) == 0 {
		m.AuditConf.PregnantH = path
	}
	path = os.Getenv(SETU_SPECIAL_H_PATH)
	if len(path) > 0 && len(m.AuditConf.SpecialH) == 0 {
		m.AuditConf.SpecialH = path
	}
	path = os.Getenv(SETU_HIPS_H_PATH)
	if len(path) > 0 && len(m.AuditConf.HipsH) == 0 {
		m.AuditConf.HipsH = path
	}
	path = os.Getenv(SETU_FEET_H_PATH)
	if len(path) > 0 && len(m.AuditConf.FeetH) == 0 {
		m.AuditConf.FeetH = path
	}
	path = os.Getenv(SETU_CROTCH_H_PATH)
	if len(path) > 0 && len(m.AuditConf.CrotchH) == 0 {
		m.AuditConf.CrotchH = path
	}
	path = os.Getenv(SETU_INTIMATE_H_PATH)
	if len(path) > 0 && len(m.AuditConf.IntimateH) == 0 {
		m.AuditConf.IntimateH = path
	}
	path = os.Getenv(SETU_ANIME_INTIMATE_H_PATH)
	if len(path) > 0 && len(m.AuditConf.AnimeIntimateH) == 0 {
		m.AuditConf.AnimeIntimateH = path
	}

	return nil
}
