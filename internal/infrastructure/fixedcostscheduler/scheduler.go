package fixedcostscheduler

import (
	"context"
	"github.com/robfig/cron/v3"
	"log"
	"sync"
	"time"
	_ "time/tzdata"
)

const Schedule = "CRON_TZ=Asia/Shanghai 55 23 * * *"

type Scheduler struct {
	cron   *cron.Cron
	job    func(context.Context, time.Time) (int, error)
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
	mu     sync.Mutex
	zone   *time.Location
}

func New(job func(context.Context, time.Time) (int, error)) (*Scheduler, error) {
	zone, e := time.LoadLocation("Asia/Shanghai")
	if e != nil {
		return nil, e
	}
	return &Scheduler{cron: cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger), cron.SkipIfStillRunning(cron.DefaultLogger))), job: job, zone: zone}, nil
}

// Cutoff includes today only once the nightly posting time has arrived.
func Cutoff(now time.Time) time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if now.Hour() < 23 || (now.Hour() == 23 && now.Minute() < 55) {
		return today.AddDate(0, 0, -1)
	}
	return today
}
func (s *Scheduler) Start(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	s.done = make(chan struct{})
	run := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if ctx.Err() != nil {
			return
		}
		jobCtx, stop := context.WithTimeout(ctx, 5*time.Minute)
		defer stop()
		if _, e := s.job(jobCtx, Cutoff(time.Now().In(s.zone))); e != nil && ctx.Err() == nil {
			log.Printf("fixed-cost scheduler: %v", e)
		}
	}
	if _, e := s.cron.AddFunc(Schedule, run); e != nil {
		cancel()
		close(s.done)
		return e
	}
	s.cron.Start()
	log.Printf("fixed-cost scheduler started: %s", Schedule)
	// Recover missed occurrences after downtime, without posting today's bill early.
	go func() {
		defer close(s.done)
		cron.NewChain(cron.Recover(cron.DefaultLogger)).Then(cron.FuncJob(run)).Run()
	}()
	return nil
}
func (s *Scheduler) Stop() {
	s.once.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		done := s.cron.Stop()
		<-done.Done()
		if s.done != nil {
			<-s.done
		}
	})
}
