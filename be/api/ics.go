package api

import (
	"be/common"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ICSAPi struct{}

func (i *ICSAPi) Handler(g gin.IRoutes) {
	g.GET("", i.Calendar)
	g.GET("/", i.Calendar)
}

func (i *ICSAPi) Calendar(c *gin.Context) {
	start, end := common.GetWeekRange(time.Now())
	ml, err := common.GetICS(start, end, 0, 1000)
	if err != nil {
		c.Abort()
		return
	}
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
	c.Data(http.StatusOK, "text/calendar", []byte(generateICS(ml)))
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
