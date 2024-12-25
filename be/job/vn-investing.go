package job

import (
	"be/common"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

func Crawl() ([]common.EconomicEvent, error) {
	url := "https://vn.investing.com/economic-calendar/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// log.Fatalf("Failed to create request: %v", err)
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("DNT", "1")
	req.Header.Add("Sec-GPC", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// log.Fatalf("Failed to fetch URL: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// log.Fatalf("Failed with status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("get %s with status %d", url, resp.StatusCode)
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		// log.Fatalf("Failed to parse HTML: %v", err)
		return nil, err
	}

	var events []common.EconomicEvent

	// Find the tbody element with the 'pageStartAt' attribute
	doc.Find("tbody[pageStartAt]").Find("tr").Each(func(i int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() >= 6 {
			event := common.EconomicEvent{
				// Time:     strings.TrimSpace(cells.Eq(0).Text()),
				Currency: strings.TrimSpace(cells.Eq(1).Text()),
				Actual:   strings.TrimSpace(cells.Eq(3).Text()),
				Forecast: strings.TrimSpace(cells.Eq(4).Text()),
				Previous: strings.TrimSpace(cells.Eq(5).Text()),
			}
			event.Event, _ = cells.Eq(2).Attr("data-img_key")
			if event.Currency != "USD" {
				return
			}
			event.Event = strings.TrimSpace(event.Event)
			if event.Event != "bull3" {
				return
			}
			dateTime, _ := row.Attr("data-event-datetime")
			event.DateTime = strings.TrimSpace(dateTime)
			event.TimeUnix = dateTimeStringToTime(event.DateTime).Unix()
			events = append(events, event)
		}
	})

	return events, nil
}

func JobCrawlInvestingCalendar() error {
	events, err := Crawl()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("craw failed")
		return err
	}
	for _, event := range events {
		ics := common.ICS{
			Uid:       common.QuickMd5([]byte(fmt.Sprintf("%s_%d", event.Actual, event.TimeUnix))),
			Name:      event.Actual,
			Desp:      fmt.Sprintf("%s Dự báo(%s) Trước đó(%s)", event.Actual, event.Forecast, event.Previous),
			StartUnix: event.TimeUnix,
			EndUnix:   event.TimeUnix + 30*60,
		}
		err = ics.Insert()
		if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			zap.L().With(zap.Error(err)).Error("insert new ics failed")
			return nil
		}
	}
	logEvent := common.CrawlLog{
		EventId:  common.CrawlLogEventIdInvestingCalendar,
		DataObjs: events,
	}
	if err := logEvent.Insert(); err != nil {
		zap.L().With(zap.Error(err)).Error("insert log craw failed")
	}
	return nil
}
func JobCrawAndImportEventInvestingCalendar(ctx context.Context) {
	lastRunSucces := time.Time{}
	fnCrawAndImport := func() {
		events, err := Crawl()
		if err != nil {
			zap.L().With(zap.Error(err)).Error("craw failed")
			return
		}
		for _, event := range events {
			ics := common.ICS{
				Uid:       common.QuickMd5([]byte(fmt.Sprintf("%s_%d", event.Actual, event.TimeUnix))),
				Name:      event.Actual,
				Desp:      fmt.Sprintf("%s Dự báo(%s) Trước đó(%s)", event.Actual, event.Forecast, event.Previous),
				StartUnix: event.TimeUnix,
				EndUnix:   event.TimeUnix + 30*60,
			}
			err = ics.Insert()
			if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				zap.L().With(zap.Error(err)).Error("insert new ics failed")
				return
			}
		}
		logEvent := common.CrawlLog{
			DataObjs: events,
		}
		if err := logEvent.Insert(); err != nil {
			zap.L().With(zap.Error(err)).Error("insert log craw failed")
		}
		lastRunSucces = time.Now()
	}
	fnCrawAndImport()
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			zap.L().Info("exist job craw investing calendar")
			return
		case <-ticker.C:
			now := time.Now()
			if now.Day() != lastRunSucces.Day() {
				fnCrawAndImport()
				return
			}
			if now.Add(-6 * time.Hour).Before(lastRunSucces) {
				return
			}
			fnCrawAndImport()
		}
	}
}

func dateTimeStringToTime(dateStr string) *time.Time {
	layout := "2006/01/02 15:04:05"
	location, err := time.LoadLocation("UTC") // GMT+7 corresponds to Asia/Bangkok
	if err != nil {
		fmt.Printf("Error loading location: %v\n", err)
		return &time.Time{}
	}

	// Parse the string into time.Time in the specified location
	parsedTime, err := time.ParseInLocation(layout, dateStr, location)
	if err != nil {
		fmt.Printf("Error parsing time: %v\n", err)
		return &time.Time{}
	}
	return &parsedTime

}
