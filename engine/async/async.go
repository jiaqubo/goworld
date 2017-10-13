package async

import (
	"sync"

	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/netutil"
	"github.com/xiaonanln/goworld/engine/post"
)

type AsyncCallback func(res interface{}, err error)

func (ac AsyncCallback) Callback(res interface{}, err error) {
	if ac != nil {
		post.Post(func() {
			ac(res, err)
		})
	}
}

type AsyncRoutine func() (res interface{}, err error)

type AsyncJobWorker struct {
	jobQueue chan asyncJobItem
}

type asyncJobItem struct {
	routine  AsyncRoutine
	callback AsyncCallback
}

func newAsyncJobWorker() *AsyncJobWorker {
	ajw := &AsyncJobWorker{
		jobQueue: make(chan asyncJobItem, consts.ASYNC_JOB_QUEUE_MAXLEN),
	}
	go netutil.ServeForever(ajw.loop)
	return ajw
}

func (ajw *AsyncJobWorker) appendJob(routine AsyncRoutine, callback AsyncCallback) {
	ajw.jobQueue <- asyncJobItem{routine, callback}
}

func (ajw *AsyncJobWorker) loop() {
	for item := range ajw.jobQueue {
		res, err := item.routine()
		if item.callback != nil {
			post.Post(func() {
				item.callback(res, err)
			})
		}
	}
}

var (
	asyncJobWorkersLock sync.RWMutex
	asyncJobWorkers     = map[string]*AsyncJobWorker{}
)

func getAsyncJobWorker(group string) (ajw *AsyncJobWorker) {
	asyncJobWorkersLock.RLock()
	ajw = asyncJobWorkers[group]
	asyncJobWorkersLock.RUnlock()

	if ajw == nil {
		asyncJobWorkersLock.Lock()
		ajw = asyncJobWorkers[group]
		if ajw == nil {
			ajw = newAsyncJobWorker()
			asyncJobWorkers[group] = ajw
		}
		asyncJobWorkersLock.Unlock()
	}
	return
}

func NewAsyncJob(group string, routine AsyncRoutine, callback AsyncCallback) {
	ajw := getAsyncJobWorker(group)
	ajw.appendJob(routine, callback)
}
