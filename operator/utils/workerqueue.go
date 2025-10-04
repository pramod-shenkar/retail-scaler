package utils

import (
	"context"
	"errors"

	"k8s.io/client-go/util/workqueue"
)

type workerQueue struct {
	workers, pieces int
	err             chan error
	finalErr        chan error
}

func NewWorkerQueue(workers, pieces int) *workerQueue {
	wq := &workerQueue{
		min(workers, pieces), pieces,
		make(chan error, pieces),
		make(chan error),
	}

	go func() {
		var err error
		for e := range wq.err {
			err = errors.Join(err, e)
		}
		wq.finalErr <- err
		close(wq.finalErr)
	}()

	return wq
}

func (w *workerQueue) PushErr(err error) {
	w.err <- err
}

func (w *workerQueue) Parallelize(ctx context.Context, doWorkPiece workqueue.DoWorkPieceFunc, opts ...workqueue.Options) error {
	workqueue.ParallelizeUntil(ctx, w.workers, w.pieces, doWorkPiece, opts...)
	close(w.err)
	return <-w.finalErr
}


