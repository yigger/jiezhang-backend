package statement

import "testing"

func TestStatementEffect(t *testing.T) {
	for _, tc := range []struct {
		kind           string
		source, target float64
		hasTarget      bool
	}{
		{"income", 25, 0, false}, {"expend", -25, 0, false}, {"transfer", -25, 25, true}, {"repayment", -25, -25, true},
		{"loan_in", 25, 0, false}, {"loan_out", -25, 0, false}, {"reimburse", -25, 0, false}, {"payment_proxy", -25, 0, false}, {"unknown", 0, 0, false},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			got := statementEffect(tc.kind, 25)
			if got.Source != tc.source || got.Target != tc.target || got.HasTarget != tc.hasTarget {
				t.Fatalf("effect=%+v", got)
			}
		})
	}
}
