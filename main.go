package main

import (
	"flag"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/xukgo/gsaber/compon/procUnique"
	"github.com/xukgo/gsaber/utils/fileUtil"
	execTimer "github.com/xukgo/procGuard/bllTimer"
	"github.com/xukgo/procGuard/core"
	"github.com/xukgo/procGuard/logUtil"
	"github.com/xukgo/procGuard/procRepo"
	"github.com/xukgo/procGuard/psutil"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

var (
	fhelp bool
	fversion bool
	fsearchKey string
	currentVersion = "202110281542"
)

func initFlag() {
	flag.BoolVar(&fhelp, "h", false, "this help")
	flag.BoolVar(&fversion, "v", false, "show version")
	flag.StringVar(&fsearchKey, "e", "", "search key word and list process info")

	// 改变默认的 Usage，flag包中的Usage 其实是一个函数类型。这里是覆盖默认函数实现，具体见后面Usage部分的分析
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `procGuard version: procGuard.%s
Options:
`,currentVersion)
	flag.PrintDefaults()
}

func main() {
	initFlag()
	flag.Parse()

	if fhelp {
		flag.Usage()
		return
	}

	if fversion {
		fmt.Fprintf(os.Stdout, "procGuard.%s\n",currentVersion)
		return
	}

	if len(fsearchKey) > 0 {
		procs, err := psutil.FilterGetProcCmdInfos()
		if err != nil {
			fmt.Fprintf(os.Stdout, "get process error:%s\n", err.Error())
			return
		}
		procInfos := core.GetProcess(procs, "", []string{fsearchKey}, nil)

		matchCount := 0
		if len(procInfos) > 0 {
			for _, info := range procInfos {
				exe := info.Exe
				cmdline := info.Cmdline
				fmt.Fprintf(os.Stdout, "%s  =>  %s\n", exe, cmdline)
				matchCount++
			}
		}

		fmt.Fprintf(os.Stdout, "match process count:%d\n", matchCount)
		return
	}

	var err error
	var procLocker = procUnique.NewLocker("procGuard_hms")
	err = procLocker.Lock()
	if err != nil {
		log.Println("应用不允许重复运行")
		os.Exit(-1)
	}

	defer procLocker.Unlock()

	logUtil.InitLogger()

	filePath := fileUtil.GetAbsUrl("conf/excludePrefix.xml")
	content, err := os.ReadFile(filePath)
	if err != nil {
		logUtil.Common().Error("read file error", zap.Error(err))
		return
	}
	excludeConf := new(core.ExcludeCommandXmlRoot)
	err = excludeConf.FillWithXml(string(content))
	if err != nil {
		logUtil.Common().Error("ExcludeCommand unmarshal error", zap.Error(err))
		return
	}

	filePath = fileUtil.GetAbsUrl("conf/crond.yml")
	content, err = os.ReadFile(filePath)
	if err != nil {
		logUtil.Common().Error("read file error", zap.Error(err))
		return
	}

	conf := new(core.ProjectConfig)
	err = conf.FillWithYaml(content)
	if err != nil {
		logUtil.Common().Error("ProjectConfig unmarshal error", zap.Error(err))
		return
	}

	for _, guard := range conf.ProcGuards {
		if !guard.Enable {
			continue
		}
		logUtil.Common().Info(guard.ToDescription())
	}

	//go func() {
	//	logUtil.Common().Info("StartProfWebService")
	//	err := http.ListenAndServe(":60044", nil)
	//	if err != nil {
	//		logUtil.Common().Error("StartProfWebService error", zap.Error(err))
	//	}
	//}()

	go startTimers(conf)

	for {
		time.Sleep(time.Hour)
	}
}

func startTimers(conf *core.ProjectConfig) {
	uuid.NewV1()
	execTimer.Start()

	tasks := conf.TimerTasks
	for _, task := range tasks {
		if !task.Enable {
			continue
		}
		logUtil.Common().Info(task.ToDescription())
		logUtil.Common().Info("启用定时任务", zap.String("description", task.Description))
		execTimer.AddCronExec(time.Second,uuid.NewV1().String(),task, taskJob)
	}

	procRepo.Refresh()
	guards := conf.ProcGuards
	count := 0
	for idx, guard := range guards {
		if !guard.Enable {
			continue
		}
		logUtil.Common().Info(guard.ToDescription())
		logUtil.Common().Info("启用服务守护", zap.String("description", guard.Description))
		uid := uuid.NewV1().String()
		fmt.Println(uid)
		time.Sleep(time.Millisecond*10)
		execTimer.AddCronExec(time.Second,uid,guards[idx], guardJob)
		count++
	}
	if count > 0{
		go procRepo.Start()
	}
}

func taskJob(k interface{}, v interface{}) {
	task := v.(*core.TimerTaskConfig)
	task.Action()
}

func guardJob(k interface{}, v interface{}) {
	procInfos := procRepo.GetProcInfos()
	if len(procInfos) == 0{
		return
	}

	guard := v.(*core.ProcGuardConfig)
	go guard.CheckAndDo(procInfos)
}

//func holdAllEnable(execArr []*core.ProcGuardConfig) {
//	if len(execArr) == 0 {
//		return
//	}
//
//	var err error
//	var procInfos []*psutil.ProcCmdInfo
//	var getProcTime time.Time
//	pidArr := make([]int, len(execArr), len(execArr))
//	for i := 0; i < len(pidArr); i++ {
//		pidArr[i] = -1
//	}
//
//	for {
//		procInfos = nil
//
//		for idx, guard := range execArr {
//			if pidArr[idx] > 0 && psutil.CheckPidExist(pidArr[idx]) {
//				continue
//			}
//
//			if procInfos == nil {
//				procInfos, err = psutil.FilterGetProcCmdInfos()
//				if err != nil {
//					logUtil.Common().Error("get process error", zap.Error(err))
//					time.Sleep(time.Second * 5)
//					continue
//				}
//				getProcTime = time.Now()
//			}
//
//			pid := guard.CheckAndDo(getProcTime, procInfos)
//			if pid > 0 {
//				pidArr[idx] = pid
//			}
//		}
//		time.Sleep(time.Second * 2)
//	}
//}
