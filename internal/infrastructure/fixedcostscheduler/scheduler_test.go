package fixedcostscheduler

import (
	"github.com/robfig/cron/v3"
	"testing"
	"time"
)

func TestNightlyScheduleAndStartupCutoff(t *testing.T) {
	s, e := cron.ParseStandard(Schedule)
	if e != nil {
		t.Fatal(e)
	}
	zone, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Date(2026, 10, 7, 23, 54, 0, 0, zone)
	if got := s.Next(now); got.Hour() != 23 || got.Minute() != 55 || got.Day() != 7 {
		t.Fatal(got)
	}
	if got := Cutoff(now).Format("2006-01-02"); got != "2026-10-06" {
		t.Fatal(got)
	}
	if got := Cutoff(now.Add(time.Minute)).Format("2006-01-02"); got != "2026-10-07" {
		t.Fatal(got)
	}
}
