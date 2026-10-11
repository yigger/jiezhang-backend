package statement

import (
	"github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
)

func recordEffect(record model.Statement) repo.BalanceEffect {
	if record.BalanceApplied != nil && !*record.BalanceApplied {
		return repo.BalanceEffect{}
	}
	return statementEffect(record.Type, record.Amount)
}

func statementEffect(kind string, amount float64) repo.BalanceEffect {
	switch kind {
	case "income", "loan_in":
		return repo.BalanceEffect{Source: amount}
	case "transfer":
		return repo.BalanceEffect{Source: -amount, Target: amount, HasTarget: true}
	case "repayment":
		return repo.BalanceEffect{Source: -amount, Target: -amount, HasTarget: true}
	case "expend", "loan_out", "reimburse", "payment_proxy":
		return repo.BalanceEffect{Source: -amount}
	default:
		return repo.BalanceEffect{}
	}
}
