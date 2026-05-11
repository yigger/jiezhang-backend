package helper

import (
	"fmt"
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
)

func StatementTitle(row repository.StatementListRowRecord) string {
	if row.Type == "transfer" || row.Type == "repayment" {
		return fmt.Sprintf("%s->%s", row.AssetName, row.TargetAssetName)
	}
	return ""
}

func WeekdayCN(wd time.Weekday) string {
	switch wd {
	case time.Sunday:
		return "周日"
	case time.Monday:
		return "周一"
	case time.Tuesday:
		return "周二"
	case time.Wednesday:
		return "周三"
	case time.Thursday:
		return "周四"
	case time.Friday:
		return "周五"
	case time.Saturday:
		return "周六"
	default:
		return ""
	}
}
