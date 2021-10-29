package core

import (
	"fmt"
	"github.com/xukgo/procGuard/logUtil"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type ProjectConfig struct {
	TimerTasks []*TimerTaskConfig `yaml:"TimerTask"`
	ProcGuards []*ProcGuardConfig `yaml:"ProcGuard"`
}

func (this *ProjectConfig) FillWithYaml(data []byte) error {
	err := yaml.Unmarshal(data, this)
	if err != nil {
		logUtil.Common().Error("ProjectConfig unmarshal yaml error", zap.Error(err))
		return err
	}
	for idx := range this.TimerTasks {
		this.TimerTasks[idx].AfterStartUp = this.TimerTasks[idx].StartupDelay <= 0
		this.TimerTasks[idx].IntervalSec, err = ParseInterval(this.TimerTasks[idx].IntervalStr)
		if err != nil {
			return err
		}
		if !this.TimerTasks[idx].CheckParam() {
			return fmt.Errorf("TimerTask校验参数未通过")
		}
	}
	for idx := range this.ProcGuards {
		this.ProcGuards[idx].IntervalSec, err = ParseInterval(this.ProcGuards[idx].IntervalStr)
		if err != nil {
			return err
		}
		if !this.ProcGuards[idx].CheckParam() {
			return fmt.Errorf("RuleExecConfig校验参数未通过")
		}
	}
	return nil
}
