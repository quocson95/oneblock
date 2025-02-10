package job

import (
	"be/common"
	"context"
	"runtime"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type JobDesp struct {
	Fn             func() error
	Interval       time.Duration
	LastRunSuccess time.Time
}

func StartJob(ctx context.Context) {
	jobs := make(map[common.CrawlLogEventId]*JobDesp)
	jobs[common.CrawlLogEventIdInvestingCalendar] = &JobDesp{
		Fn:       JobCrawlInvestingCalendar,
		Interval: 12 * time.Hour,
	}
	dataTrading := new(CrawlDataTrading)
	jobs[common.CrawlLogEventIdDataTradingBtcGold] = &JobDesp{
		Fn:       dataTrading.BtcGoldCore,
		Interval: 12 * time.Hour,
	}
	jobs[common.CrawlLogEventIdDataTradingSP500] = &JobDesp{
		Fn:       dataTrading.Sp500,
		Interval: 12 * time.Hour,
	}
	jobs["expire_subscribe"] = &JobDesp{
		Fn:       JobExpireSubscribe,
		Interval: 1 * time.Hour,
	}
	ticker := time.NewTicker(1 * time.Minute)
	for _, v := range jobs {
		err := v.Fn()
		if err == nil {
			v.LastRunSuccess = time.Now()
		}

	}
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))
	for {
		select {
		case <-ctx.Done():
			zap.L().Info("exist job craw investing calendar")
			return
		case <-ticker.C:
			now := time.Now()
			for jobID, jobDesp := range jobs {
				_ = jobID
				lastRunSuccess := jobDesp.LastRunSuccess
				if now.Day() != lastRunSuccess.Day() || now.Add(-6*time.Hour).After(lastRunSuccess) {
					g.Go(func(j *JobDesp) func() error {
						return func() error {
							err := j.Fn()
							if err == nil {
								j.LastRunSuccess = time.Now()
							}
							return nil
						}
					}(jobDesp))
					continue
				}

			}
		}
	}

}
