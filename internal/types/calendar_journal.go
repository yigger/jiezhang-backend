package types

type CalendarJournalInput struct {
	Date        string `json:"date"`
	Mood        string `json:"mood"`
	Note        string `json:"note"`
	ZeroExpense bool   `json:"zero_expense"`
}
type CalendarJournalMonth struct {
	Entries      []CalendarJournalInput `json:"entries"`
	RevokedDates []string               `json:"revoked_dates"`
}
