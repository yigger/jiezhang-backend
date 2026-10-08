package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm/clause"
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

type statementMutation struct{ *StatementRepository }

func (r *StatementRepository) WithinTransaction(ctx context.Context, fn func(repo.Mutation) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(&statementMutation{&StatementRepository{db: tx}}) })
}
func (r *statementMutation) LockCurrent(ctx context.Context, id, bookID int64) (tablemodel.Statement, error) {
	return r.getStatementForUpdate(r.db.WithContext(ctx), id, bookID)
}

func (r *statementMutation) AttachProject(ctx context.Context, book, statement, project, consumer, user int64) error {
	var p tablemodel.InsightProject
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND account_book_id = ?", project, book).Take(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return repo.ErrInsightRecordNotFound
		}
		return err
	}
	if p.Archived || consumer <= 0 {
		return repo.ErrInsightRecordNotFound
	}
	var members []tablemodel.AccountBookCollaborator
	if err := r.db.Where("account_book_id = ? AND user_id IN ?", book, []int64{user, consumer}).Find(&members).Error; err != nil {
		return err
	}
	var b tablemodel.AccountBook
	if err := r.db.Where("id = ?", book).Take(&b).Error; err != nil {
		return err
	}
	validUser, validConsumer := b.UserID == user, b.UserID == consumer
	for _, m := range members {
		if m.UserID == user {
			validUser = true
		}
		if m.UserID == consumer {
			validConsumer = true
		}
	}
	if !validUser || !validConsumer {
		return repo.ErrInsightRecordNotFound
	}
	var participants []int64
	if len(p.ParticipantIDs) > 0 {
		if err := json.Unmarshal(p.ParticipantIDs, &participants); err != nil {
			return err
		}
	}
	if len(participants) > 0 {
		found := false
		for _, id := range participants {
			if id == consumer {
				found = true
			}
		}
		if !found {
			return repo.ErrInsightRecordNotFound
		}
	}
	now := time.Now()
	a := tablemodel.InsightStatementAnnotation{AccountBookID: book, CreatorID: user, StatementID: statement, ProjectID: &project, ConsumerID: &consumer, CreatedAt: now, UpdatedAt: now}
	return r.db.Create(&a).Error
}
