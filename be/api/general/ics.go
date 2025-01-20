package api

import (
	"be/common"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type ICSAPi struct{}

func (i *ICSAPi) Handler(g *echo.Group) {
	g.GET("", i.Calendar)
	g.GET("/", i.Calendar)
	g.GET("/register", i.CalendarRegister)
	g.GET("/ios", i.Calendar)
	g.GET("/cal", i.Calendar)
}

func (i *ICSAPi) CalendarRegister(c echo.Context) error {
	userAgent := c.Request().UserAgent()
	if isIOS(userAgent) {
		return c.Redirect(http.StatusMovedPermanently, "webcal://api.oneblock.vn/be/ics/ios")
	}
	return c.Redirect(http.StatusMovedPermanently, "https://api.oneblock.vn/be/ics/cal")
}

func (i *ICSAPi) Calendar(c echo.Context) error {
	cacheData, exist := common.CacheDataPool.Get("calendar")
	var calData []byte
	if exist && cacheData.InvalidAt.Before(time.Now()) {
		calData = cacheData.Data
	} else {
		start, end := common.GetWeekRange(time.Now())
		start = start.Add(-7 * 24 * time.Hour)
		end = end.Add(7 * 24 * time.Hour)
		ml, err := common.GetICS(start, end, 0, 1000)
		if err != nil {
			return c.NoContent(http.StatusOK)
		}
		calData = []byte(generateICS(ml))
		common.CacheDataPool.Add("calendar", common.CacheData{
			Data:      calData,
			Mime:      "text/calendar",
			InvalidAt: time.Now().Add(10 * time.Minute),
		})
	}
	resp := c.Response()
	resp.Writer.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
	c.SetResponse(resp)
	return c.Blob(http.StatusOK, "text/calendar", calData)
}

func generateICS(ml []common.ICS) string {
	builder := strings.Builder{}
	// Generate ICS content
	// icsContent := "BEGIN:VCALENDAR\n"
	// icsContent += "VERSION:2.0\n"
	// icsContent += "PRODID:-//Oneblock//NONSGML v1.0//EN\n"
	// icsContent += "CALSCALE:GREGORIAN\n"
	builder.WriteString("BEGIN:VCALENDAR\n")
	builder.WriteString("VERSION:2.0\n")
	builder.WriteString("PRODID:-//Oneblock//NONSGML v1.0//EN\n")
	builder.WriteString("X-WR-CALNAME:Oneblock Lịch Kinh Tế\n")
	builder.WriteString("CALSCALE:GREGORIAN\n")
	for _, event := range ml {
		// Event details
		startTime := time.Unix(event.StartUnix, 0)
		if event.EndUnix < event.StartUnix {
			event.EndUnix = event.StartUnix + 3600
		}
		endTime := time.Unix(event.StartUnix, 0)
		builder.WriteString("BEGIN:VEVENT\n")
		builder.WriteString(fmt.Sprintf("UID:%s@oneblock.vn\n", event.Uid))
		builder.WriteString(fmt.Sprintf("DTSTAMP:%s\n", startTime.Format("20060102T150405Z")))
		builder.WriteString((fmt.Sprintf("DTSTART:%s\n", startTime.Format("20060102T150000Z"))))
		builder.WriteString(fmt.Sprintf("DTEND:%s\n", endTime.Format("20060102T150000Z")))
		builder.WriteString(fmt.Sprintf("SUMMARY:%s\n", event.Name))
		builder.WriteString(fmt.Sprintf("DESCRIPTION:%s.\n", event.Desp))
		builder.WriteString("LOCATION:Online\n")
		builder.WriteString("END:VEVENT\n")
	}

	// iCalendar footer
	builder.WriteString("END:VCALENDAR\n")
	return builder.String()
}

func isIOS(userAgent string) bool {
	// Check for iOS-related terms in the User-Agent string
	return strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad") || strings.Contains(userAgent, "iPod")
}
