package core

import (
	"bytes"
	"fmt"
	"github.com/xukgo/procGuard/logUtil"
	"github.com/xukgo/procGuard/psutil"
	"go.uber.org/zap"
	"time"
)

type ProcGuardConfig struct {
	Enable        bool              `yaml:"enable"`
	StartupDelay  int64             `yaml:"startupDelay"`
	Description   string            `yaml:"description"`
	IntervalStr   string            `yaml:"interval"`
	IntervalSec   int64             `yaml:"-"`
	Commands      []string          `yaml:"command"`
	CheckConfig   *CheckExistConfig `yaml:"check"`
	LastCheckTime time.Time         `yaml:"-"`
	Pid           int               `yaml:"-"`
}

type CheckExistConfig struct {
	ExecPath   string   `yaml:"execPath"`
	IncludeCmd []string `yaml:"includeCmd"`
	ExcludeCmd []string `yaml:"excludeCmd"`
}

func (this *ProcGuardConfig) ToDescription() string {
	var bf bytes.Buffer
	bf.WriteString("解析启用服务守护\n")
	bf.WriteString(fmt.Sprintf("    Description=>%s\n", this.Description))
	bf.WriteString(fmt.Sprintf("    StartupDelay=>%d秒\n", this.StartupDelay))
	bf.WriteString(fmt.Sprintf("    Interval=>%s\n", this.IntervalStr))
	bf.WriteString(fmt.Sprintf("    Commands:\n"))
	for _, cmd := range this.Commands {
		bf.WriteString(fmt.Sprintf("        =>%s\n", cmd))
	}
	return bf.String()
}

func (this *ProcGuardConfig) CheckParam() bool {
	if this.StartupDelay < 0 {
		logUtil.Common().Error("ProcGuardConfig StartupDelay is not valid")
		return false
	}
	if this.IntervalSec <= 0 {
		logUtil.Common().Error("ProcGuardConfig interval is not valid")
		return false
	}
	if len(this.Commands) == 0 {
		logUtil.Common().Error("ProcGuardConfig commands is not valid")
		return false
	}
	if this.CheckConfig == nil {
		logUtil.Common().Error("ProcGuardConfig check exist config is not valid")
		return false
	}
	if len(this.CheckConfig.ExecPath) == 0 && len(this.CheckConfig.IncludeCmd) == 0 {
		logUtil.Common().Error("ProcGuardConfig check exist config not allow include param and exec path both are empty")
		return false
	}
	return true
}

func (this *ProcGuardConfig) CheckAndDo(procInfos []*psutil.ProcCmdInfo) {
	if this.StartupDelay > 0 {
		duration, err := psutil.GetUptime()
		if err != nil {
			logUtil.Common().Error("GetSystemUptime error", zap.Error(err))
			return
		}
		if duration.Seconds() < float64(this.StartupDelay) {
			return
		}
	}

	if time.Since(this.LastCheckTime).Seconds() < float64(this.IntervalSec) {
		return
	}
	this.LastCheckTime = time.Now()

	if this.Pid > 0 && psutil.CheckPidExist(this.Pid) {
		return
	}
	this.Pid = -1

	infos := GetProcess(procInfos, this.CheckConfig.ExecPath, this.CheckConfig.IncludeCmd, this.CheckConfig.ExcludeCmd)
	if len(infos) == 1 {
		this.Pid = infos[0].Pid
		return
	}
	if len(infos) > 1 {
		logUtil.Common().Warn("get process multi count", zap.Int("count", len(infos)), zap.String("desc", this.Description))
		for _, v := range infos {
			logUtil.Common().Warn("multi command", zap.String("command", v.Cmdline), zap.String("exe", v.Exe), zap.Int("pid", v.Pid))
		}
		return
	}

	for _, cmd := range this.Commands {
		logUtil.Common().Info("服务守护开始执行", zap.String("description", this.Description), zap.String("cmd", cmd))
		outStr, errStr, err := ExecCmdline(cmd)
		if err != nil {
			logUtil.Common().Error("服务守护执行失败", zap.Error(err), zap.String("description", this.Description),
				zap.String("cmd", cmd), zap.String("stdout", outStr), zap.String("stderr", errStr))
		} else {
			logUtil.Common().Info("服务守护执行成功", zap.Error(err), zap.String("description", this.Description),
				zap.String("cmd", cmd), zap.String("stdout", outStr))
		}
		time.Sleep(time.Millisecond*30)
	}
}
