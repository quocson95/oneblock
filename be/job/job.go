package job

import (
	"be/common"
	"context"
	"time"

	"go.uber.org/zap"
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
	ticker := time.NewTicker(1 * time.Minute)
	for _, v := range jobs {
		err := v.Fn()
		if err == nil {
			v.LastRunSuccess = time.Now()
		}

	}
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
					go func() {
						err := jobDesp.Fn()
						if err == nil {
							jobDesp.LastRunSuccess = time.Now()
						}
					}()
					continue
				}

			}
		}
	}

}
