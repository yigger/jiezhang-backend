package mysql

import (
	"time"

	"github.com/yigger/jiezhang-backend/internal/repository"
	"gorm.io/gorm"
)

// createAccountBookInTx creates an account book with collaborators, categories,
// and assets inside the given transaction. Returns the created book row.
func createAccountBookInTx(tx *gorm.DB, input repository.AccountBookCreateInput) (accountBookRow, error) {
	now := time.Now()
	created := accountBookRow{
		UserID:      input.UserID,
		AccountType: input.AccountType,
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := tx.Table("account_books").Create(&created).Error; err != nil {
		return accountBookRow{}, err
	}

	collaborator := accountBookCollaboratorRow{
		AccountBookID: created.ID,
		UserID:        input.UserID,
		Role:          "owner",
		Remark:        input.UserNickname,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := tx.Table("account_book_collaborators").Create(&collaborator).Error; err != nil {
		return accountBookRow{}, err
	}

	for statementType, parents := range input.Categories {
		for parentOrder, parent := range parents {
			parentCategory := categoryRow{
				UserID:        input.UserID,
				AccountBookID: created.ID,
				Name:          parent.Name,
				IconPath:      parent.IconPath,
				Type:          statementType,
				Order:         parentOrder + 1,
				ParentID:      0,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Table("categories").Create(&parentCategory).Error; err != nil {
				return accountBookRow{}, err
			}
			for childOrder, child := range parent.Childs {
				childCategory := categoryRow{
					UserID:        input.UserID,
					AccountBookID: created.ID,
					Name:          child.Name,
					IconPath:      child.IconPath,
					Type:          statementType,
					Order:         childOrder,
					ParentID:      parentCategory.ID,
					CreatedAt:     now,
					UpdatedAt:     now,
				}
				if err := tx.Table("categories").Create(&childCategory).Error; err != nil {
					return accountBookRow{}, err
				}
			}
		}
	}

	for _, asset := range input.Assets {
		assetType := asset.Type
		if assetType == "" {
			assetType = "deposit"
		}
		parentAsset := assetRow{
			AccountBookID: created.ID,
			Name:          asset.Name,
			IconPath:      asset.IconPath,
			Type:          assetType,
			ParentID:      0,
			CreatorID:     input.UserID,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Table("assets").Create(&parentAsset).Error; err != nil {
			return accountBookRow{}, err
		}
		for _, child := range asset.Childs {
			childAsset := assetRow{
				AccountBookID: created.ID,
				Name:          child.Name,
				IconPath:      child.IconPath,
				Type:          assetType,
				ParentID:      parentAsset.ID,
				CreatorID:     input.UserID,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Table("assets").Create(&childAsset).Error; err != nil {
				return accountBookRow{}, err
			}
		}
	}

	return created, nil
}
