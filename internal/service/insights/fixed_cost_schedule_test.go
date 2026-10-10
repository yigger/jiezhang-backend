package insights

import (
	"testing"
	"time"
)

func TestFixedCostMonthEndsRestoreRequestedDay(t *testing.T) {
	d := time.Date(2026, 1, 31, 0, 0, 0, 0, fixedCostZone)
	for _, want := range []string{"2026-02-28", "2026-03-31", "2026-04-30", "2026-05-31"} {
		d = nextFixedCostDate(d, 1, 31)
		if d.Format("2006-01-02") != want {
			t.Fatalf("got %s want %s", d, want)
		}
	}
	leap := nextFixedCostDate(time.Date(2028, 1, 31, 0, 0, 0, 0, fixedCostZone), 1, 31)
	if leap.Day() != 29 {
		t.Fatal(leap)
	}
	quarterly := nextFixedCostDate(time.Date(2026, 11, 30, 0, 0, 0, 0, fixedCostZone), 3, 31)
	if quarterly.Format("2006-01-02") != "2027-02-28" {
		t.Fatal(quarterly)
	}
}
func TestFixedCostStartsAtNextEligibleDate(t *testing.T) {
	now := time.Date(2026, 10, 7, 15, 0, 0, 0, fixedCostZone)
	for _, c := range []struct {
		interval, day int
		want          string
	}{{1, 7, "2026-10-07"}, {1, 1, "2026-11-01"}, {3, 1, "2027-01-01"}, {1, 31, "2026-10-31"}} {
		d := firstFixedCostDate(now, c.interval, c.day)
		if d.Format("2006-01-02") != c.want {
			t.Fatalf("%+v got %s", c, d)
		}
	}
	utc := time.Date(2026, 10, 6, 20, 0, 0, 0, time.UTC)
	if firstFixedCostDate(utc, 1, 7).Format("2006-01-02") != "2026-10-07" {
		t.Fatal("timezone lost")
	}
}

func TestFixedCostDatesUseCalendarDaysForStorage(t *testing.T) {
	due := firstFixedCostDate(time.Date(2026, 10, 7, 0, 30, 0, 0, fixedCostZone), 1, 7)
	if due.Location() != time.Local || due.Format("2006-01-02") != "2026-10-07" {
		t.Fatal("date changed with server timezone", due)
	}
}
