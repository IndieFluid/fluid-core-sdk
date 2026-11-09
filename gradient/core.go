package gradient

import (
	"sync/atomic"
	"time"
)

type core struct {
	alpha         float64
	actualRTT     uint64
	actualNoLoad  uint64
	lastTimeCheck int64
	TTL           int64
}

type Core interface {
	UpdateRTTActual(sampleRTT time.Duration)
	UpdateRTTNoload(currentRTT uint64)
	Gradient() float64
}

func NewCore(alpha float64, actualRTT uint64) Core {
	return &core{
		alpha:     alpha,
		actualRTT: actualRTT,
	}
}

func (c *core) UpdateRTTActual(sampleRTT time.Duration) {
	sample := float64(sampleRTT)
	for {
		oldVal := atomic.LoadUint64(&c.actualRTT)
		if oldVal == 0 {
			if atomic.CompareAndSwapUint64(&c.actualRTT, 0, uint64(sample)) {
				break
			}
		}

		oldFloat := float64(oldVal)
		newFloat := (c.alpha * sample) + ((1.0 - c.alpha) * oldFloat)
		newVal := uint64(newFloat)

		if atomic.CompareAndSwapUint64(&c.actualRTT, oldVal, newVal) {
			break
		}
	}
}

func (c *core) UpdateRTTNoload(currentRTT uint64) {
	now := time.Now().UnixNano()
	minTime := atomic.LoadInt64(&c.lastTimeCheck)
	minn := atomic.LoadUint64(&c.actualNoLoad)

	if currentRTT < minn || minn == 0 || now-minTime > c.TTL {
		c.actualNoLoad = currentRTT
		c.lastTimeCheck = now
	}
}

func (c *core) Gradient() float64 {
	return float64(c.actualNoLoad) / float64(c.actualRTT)
}
