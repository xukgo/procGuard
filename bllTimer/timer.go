package execTimer

import (
	"github.com/xukgo/gsaber/compon/timewheel"
	"time"
)

var tw *timewheel.TimeWheel

func Start() {
	tw = timewheel.New(time.Millisecond*250, 400, 300, nil)
	tw.Start()
}
