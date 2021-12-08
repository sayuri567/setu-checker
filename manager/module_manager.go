package manager

import (
	"github.com/sayuri567/setu-checker/constants"
	"github.com/sayuri567/setu-checker/jobs"
	"github.com/sayuri567/tool/module"
	"github.com/sayuri567/tool/module/crontab"
	"github.com/sayuri567/tool/module/logger"
	"github.com/sirupsen/logrus"
)

type ModuleManager struct {
	*module.DefaultModuleManager
	Logger  *logger.LoggerModule
	Crontab *crontab.CrontabModule
}

var Mod = &ModuleManager{
	DefaultModuleManager: module.NewDefaultModuleManager(),
}

func (m *ModuleManager) Init() error {
	logger.SetConfig(&logger.Config{Level: logrus.InfoLevel.String(), TimeFormat: constants.TimeFormat, ExtendFields: map[string]string{"@type": "setu"}})
	for _, cron := range m.getCron() {
		crontab.RegisterCron(cron)
	}

	m.Logger = m.AppendModule(logger.GetLoggerModule()).(*logger.LoggerModule)
	m.Crontab = m.AppendModule(crontab.GetCrontabModule()).(*crontab.CrontabModule)

	return m.DefaultModuleManager.Init()
}

func (m *ModuleManager) getCron() []*crontab.Crontab {
	return []*crontab.Crontab{
		{Spec: "36 */12 * * * ?", Cmd: jobs.CheckImages},
	}
}
