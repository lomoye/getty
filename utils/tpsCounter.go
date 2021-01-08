package utils

import (
	"sync"
	"time"
)

//全局默认的tps计数器
var GlobalTpsCounter = NewTpsCounter(6, 10)

type TpsCounter struct {
	time     int64   //纳秒时间戳
	count    int64   //计数器
	size     int64   //窗口数量
	index    int64   //窗口索引
	interval int64   //窗口间隔大小 单位秒
	windows  []int64 //窗口集合
	lock     sync.RWMutex
}

//初始化
func NewTpsCounter(size int64, interval int64) *TpsCounter {
	return &TpsCounter{
		time:     0,
		count:    0,
		size:     size,
		index:    0,
		interval: interval,
		windows:  make([]int64, size),
	}
}

//增加tps
func (counter *TpsCounter) Inc() {
	counter.lock.Lock()
	defer counter.lock.Unlock()

	nowTime := time.Now().UnixNano()
	if nowTime < counter.time+counter.interval*1000000000 {
		// 在时间窗口内
		counter.count++
	} else {
		counter.time = nowTime // 开启新的窗口
		counter.windows[counter.index] = counter.count
		counter.index = (counter.index + 1) % counter.size
		counter.count = 1 // 初始化计数器,由于这个请求属于当前新开的窗口，所以记录这个请求
	}
}

func (counter *TpsCounter) GetWindows() []int64 {
	counter.lock.RLock()
	defer counter.lock.RUnlock()
	return counter.windows
}
