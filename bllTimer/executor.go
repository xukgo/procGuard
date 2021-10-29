package execTimer

import (
	"time"
)

func AddUniqueDelayExec(delay time.Duration, key interface{}, data interface{}, job func(interface{}, interface{})) {
	tw.Remove(key)
	tw.AddFunc(delay, key, data, job)
}

func AddDelayExec(delay time.Duration, key interface{}, data interface{}, job func(interface{}, interface{})) {
	tw.AddFunc(delay, key, data, job)
}

func AddCronExec(delay time.Duration, key interface{}, data interface{}, job func(interface{}, interface{})) {
	tw.AddCronFunc(delay, key, data, job)
}

func Remove(key interface{}) {
	tw.Remove(key)
}
