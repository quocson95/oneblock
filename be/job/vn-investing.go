package job

import (
	"be/common"
	"context"
	"fmt"
	"io"
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

	// req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
	// req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	// req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	// req.Header.Add("DNT", "1")
	// req.Header.Add("Sec-GPC", "1")
	req.Header.Add("Cookie", "__cf_bm=pYmoFBXLrIpcYJzeBy9ZOMa.oIEn38gvRp4xqj8sI0w-1735532716-1.0.1.1-wvZlytIHPqLJjY_amryp7DmChuFiP5XoZwSJhyxAENA4fFLjOJ0HkfxwSwJV71QDyEHFQF822trM1ocyUtiRCfckNW1hvZhtQlEDLXUsbk0; adBlockerNewUserDomains=1734664053; browser-session-counted=true; firstUdid=0; ses_id=N3lkJTU6PzdiJmhuYTA5P2I2MmliZmZnPT1mZDs7NyFnczQ6YzQ3cWFuayVmZTMvZWM2P2Y4MTJgN24wYGYzZzdjZDc1Nj9kYjNoZWExOT9iZzJrYmVmZz07Zjc7Pzc4Z2g0ZWM2NzZhYGtjZj4zb2V3NipmIjEgYDJuPmAhM3Q3OGQlNWU%2FZGIzaGZhMDk7YmIyO2IwZjE9O2ZsOzs3L2cs; smd=e9bd728664efe00612ad9e26cdc156d4-1735532715; udid=e9bd728664efe00612ad9e26cdc156d4; user-browser-sessions=3; Adsfree_conversion_score=3; PHPSESSID=icfthldbdo2sfkr9tfl3l3784g; __cflb=02DiuF9qvuxBvFEb2q9Qemd3EPFFTD8S94pEkGvMTsSHr; adsFreeSalePopUp64a162f827687494b4f68a4452f3574e=1; comment_notification_245033117=1; geoC=VN; gtmFired=OK; nyxDorf=MzdkKTVmN2g3f2tlNGI0Mzd4N21lZmJn; page_equity_viewed=0; r_p_s_n=1; upa=eyJpbnZfcHJvX2Z1bm5lbCI6IiIsIm1haW5fYWMiOiIxMCIsIm1haW5fc2VnbWVudCI6IjIiLCJkaXNwbGF5X3JmbSI6IjExMyIsImFmZmluaXR5X3Njb3JlX2FjX2VxdWl0aWVzIjoiMSIsImFmZmluaXR5X3Njb3JlX2FjX2NyeXB0b2N1cnJlbmNpZXMiOiI4IiwiYWZmaW5pdHlfc2NvcmVfYWNfY3VycmVuY2llcyI6IjEiLCJhY3RpdmVfb25faW9zX2FwcCI6IjAiLCJhY3RpdmVfb25fYW5kcm9pZF9hcHAiOiIwIiwiYWN0aXZlX29uX3dlYiI6IjEiLCJpbnZfcHJvX3VzZXJfc2NvcmUiOiIwIn0%3D")

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

	return ParseCrawlInvestCal(resp.Body)
}

func ParseCrawlInvestCal(inp io.Reader) ([]common.EconomicEvent, error) {
	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(inp)
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
				// Forecast: strings.TrimSpace(cells.Eq(4).Text()),
				// Previous: strings.TrimSpace(cells.Eq(5).Text()),
			}
			event.Event, _ = cells.Eq(2).Attr("data-img_key")
			if event.Currency != "USD" {
				return
			}

			if len(event.Forecast) == 0 {
				event.Forecast = findId(row, "eventForecast")
			}
			if len(event.Previous) == 0 {
				event.Previous = findId(row, "eventPrevious")
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

func findId(doc *goquery.Selection, id string) string {
	txt := ""
	doc.Find(fmt.Sprintf(`[id^="%s_"]`, id)).Each(func(i int, selection *goquery.Selection) {
		s := strings.TrimSpace(selection.Text())
		if len(s) > 0 {
			txt = s
			return
		}
	})
	return txt
}

func JobCrawlInvestingCalendar() error {
	events, err := Crawl()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("craw failed")
		return err
	}
	InsertEventInvestCal(events)
	return nil
}

func InsertEventInvestCal(events []common.EconomicEvent) {
	var err error
	for _, event := range events {
		ics := common.ICS{
			Uid:       common.QuickMd5([]byte(fmt.Sprintf("%s_%d", event.Actual, event.TimeUnix))),
			Name:      event.Actual,
			Desp:      fmt.Sprintf("%s Dự báo(%s) Trước đó(%s)", event.Actual, event.Forecast, event.Previous),
			StartUnix: event.TimeUnix,
			EndUnix:   event.TimeUnix + 30*60,
		}
		if (ics).Exist() {
			ics.Updates(map[string]interface{}{"desp": ics.Desp})
		} else {
			err = ics.Insert()

		}
		if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			zap.L().With(zap.Error(err)).Error("insert new ics failed")
			return
		}
	}
	logEvent := common.CrawlLog{
		EventId:  common.CrawlLogEventIdInvestingCalendar,
		DataObjs: events,
	}
	if err := logEvent.Insert(); err != nil {
		zap.L().With(zap.Error(err)).Error("insert log craw failed")
	}
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
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh") // GMT+7 corresponds to Asia/Bangkok
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
