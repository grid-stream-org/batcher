package stats

type RunningAvg struct {
	count   int64
	sum     float64
	average float64
}

func NewRunningAvg() *RunningAvg {
	return &RunningAvg{}
}

func (ra *RunningAvg) Add(value float64) {
	ra.count++
	ra.sum += value
	ra.average = ra.sum / float64(ra.count)
}

func (ra *RunningAvg) Get() float64 {
	return ra.average
}

func (ra *RunningAvg) Count() int64 {
	return ra.count
}

func (ra *RunningAvg) Reset() {
	ra.count = 0
	ra.sum = 0
	ra.average = 0
}
