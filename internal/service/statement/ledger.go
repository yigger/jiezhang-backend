package statement

import "github.com/yigger/jiezhang-backend/internal/repo"

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
