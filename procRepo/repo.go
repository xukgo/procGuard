package procRepo

import (
	"github.com/xukgo/procGuard/logUtil"
	"github.com/xukgo/procGuard/psutil"
	"go.uber.org/zap"
	"sync"
	"time"
)

var singleton = initRepo()

type Repo struct {
	locker *sync.RWMutex
	lastTime time.Time
	procInfos []*psutil.ProcCmdInfo
}

func initRepo()*Repo{
	repo := new(Repo)
	repo.locker = new(sync.RWMutex)
	return repo
}

func Start(){
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		Refresh()
	}
}
func Refresh() error{
	procInfos, err := psutil.FilterGetProcCmdInfos()
	if err != nil {
		logUtil.Common().Error("get process error", zap.Error(err))
		return err
	}

	singleton.locker.Lock()
	singleton.procInfos = procInfos
	singleton.lastTime = time.Now()
	singleton.locker.Unlock()
	return nil
}

func GetProcInfos()[]*psutil.ProcCmdInfo{
	singleton.locker.RLock()
	procInfos :=singleton.procInfos
	singleton.locker.RUnlock()
	return procInfos
}