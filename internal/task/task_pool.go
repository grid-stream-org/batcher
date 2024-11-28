package task

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/destination"
	"github.com/grid-stream-org/batcher/internal/utils"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type TaskPool struct {
	cfg         *config.Pool
	tasks       chan Task
	destination destination.Destination
	dedup       *cache.Cache
	wg          sync.WaitGroup
	log         *slog.Logger
}

func NewTaskPool(
	ctx context.Context,
	cfg *config.Pool,
	dest destination.Destination,
	log *slog.Logger,
) *TaskPool {
	tp := &TaskPool{
		cfg:         cfg,
		tasks:       make(chan Task, cfg.Capacity),
		destination: dest,
		dedup:       cache.New(1*time.Minute, 5*time.Minute),
		log: log.With(
			"component", "task_pool",
			"num_workers", cfg.NumWorkers,
			"capacity", cfg.Capacity,
		),
	}

	tp.log.Info("task pool created")
	return tp
}

func (tp *TaskPool) Submit(t any) {
	task, ok := t.(Task)
	if !ok {
		tp.log.Warn("received non-task event", "type", utils.TypeOf(t))
		return
	}
	tp.submitTask(task)
}

func (tp *TaskPool) submitTask(t Task) {
	log := tp.log.With(t.LogFields()...)
	log.Debug("received task from event bus")
	if tp.dedup.Add(t.ID(), struct{}{}, 5*time.Minute) != nil {
		log.Warn("skipping duplicate task")
		return

	}
	log.Debug("submitting task")
	tp.tasks <- t
}

func (tp *TaskPool) Start(ctx context.Context) {
	tp.log.Info("starting task pool")
	for i := 0; i < tp.cfg.NumWorkers; i++ {
		tp.wg.Add(1)
		go tp.worker(ctx, i)
	}
	tp.log.Info("task pool started successfully")
}

func (tp *TaskPool) worker(ctx context.Context, workerId int) {
	defer tp.wg.Done()
	log := tp.log.With("worker_id", workerId)
	for {
		select {
		case t, ok := <-tp.tasks:
			if !ok {
				log.Debug("task channel closed, stopping worker")
				return
			}
			log := log.With(t.LogFields()...)
			log.Debug("processing task")
			outcome, err := t.Execute(workerId)
			if err != nil {
				if errors.Is(err, ErrNoDERs) {
					log.Warn("received empty DER array")
				} else {
					// For now we will just log an error if the task execution fails
					// I can't see this happening for any other reason than a bad json payload
					// Until something else arises, we can stick with this
					log.Error("task execution failed", "error", err)
				}
				continue
			}
			if err := tp.destination.Add(outcome); err != nil {
				log.Error("failed to add outcome to destination", "error", err)
				continue
			}
			log.Debug("task completed successfully")
		case <-ctx.Done():
			log.Debug("context cancelled, stopping worker", "reason", ctx.Err())
			return
		}
	}
}

func (tp *TaskPool) Wait() {
	tp.log.Info("shutting down task pool")
	close(tp.tasks)
	tp.dedup.Flush()
	tp.wg.Wait()
	tp.log.Info("task pool shutdown complete")
}

func (tp *TaskPool) LogFields() []any {
	return []any{
		"component", "task_pool",
		"num_workers", tp.cfg.NumWorkers,
		"capacity", tp.cfg.Capacity,
	}
}
