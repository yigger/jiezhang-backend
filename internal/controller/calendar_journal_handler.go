package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/yigger/jiezhang-backend/internal/service/calendarjournal"
	"github.com/yigger/jiezhang-backend/internal/types"
	"net/http"
)

type CalendarJournalHandler struct{ service *calendarjournal.Service }

func NewCalendarJournalHandler(s *calendarjournal.Service) CalendarJournalHandler {
	return CalendarJournalHandler{s}
}
func journalError(c *gin.Context, err error) {
	status, message := http.StatusInternalServerError, "手帐暂时不可用，请稍后重试"
	if errors.Is(err, calendarjournal.ErrForbidden) {
		status, message = http.StatusForbidden, err.Error()
	} else if errors.Is(err, calendarjournal.ErrInvalid) || errors.Is(err, calendarjournal.ErrFuture) || errors.Is(err, calendarjournal.ErrExpense) {
		status, message = http.StatusBadRequest, err.Error()
	}
	c.JSON(status, gin.H{"error": message})
}

// Month 获取当前用户在账簿中的月度手帐。
// @Summary 月度日历手帐
// @ID CalendarJournalHandler_Month
// @Tags CalendarJournal
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Param month query string true "YYYY-MM"
// @Success 200 {object} types.CalendarJournalMonth
// @Failure 400 {object} types.APIResponse
// @Failure 403 {object} types.APIResponse
// @Router /calendar/journal [get]
func (h CalendarJournalHandler) Month(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	result, err := h.service.Month(c.Request.Context(), book.ID, user.ID, c.Query("month"))
	if err != nil {
		journalError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Save 保存或清空某一天的个人手帐，不影响账单余额。
// @Summary 保存日历手帐
// @ID CalendarJournalHandler_Save
// @Tags CalendarJournal
// @Accept json
// @Produce json
// @Security AppID && SessionKey
// @Param account_book_id query integer false "账簿 ID"
// @Param payload body types.CalendarJournalInput true "当天手帐"
// @Success 200 {object} types.CalendarJournalInput
// @Failure 400 {object} types.APIResponse
// @Failure 403 {object} types.APIResponse
// @Router /calendar/journal [put]
func (h CalendarJournalHandler) Save(c *gin.Context) {
	book, ok := RequireAccountBook(c)
	if !ok {
		return
	}
	user, ok := RequireCurrentUser(c)
	if !ok {
		return
	}
	var in types.CalendarJournalInput
	if c.ShouldBindJSON(&in) != nil {
		journalError(c, calendarjournal.ErrInvalid)
		return
	}
	result, err := h.service.Save(c.Request.Context(), book.ID, user.ID, in)
	if err != nil {
		journalError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
