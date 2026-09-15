package mysql

import (
	"time"

	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/repo"
	"gorm.io/gorm"
)

// createAccountBookInTx creates an account book with collaborators, categories,
// and assets inside the given transaction. Returns the created book row.
func createAccountBookInTx(tx *gorm.DB, input repo.AccountBookCreateInput) (tablemodel.AccountBook, error) {
	now := time.Now()
	created := tablemodel.AccountBook{
		UserID:      input.UserID,
		AccountType: input.AccountType,
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := tx.Table("account_books").Create(&created).Error; err != nil {
		return tablemodel.AccountBook{}, err
	}

	collaborator := tablemodel.AccountBookCollaborator{
		AccountBookID: created.ID,
		UserID:        input.UserID,
		Role:          "owner",
		Remark:        input.UserNickname,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := tx.Table("account_book_collaborators").Create(&collaborator).Error; err != nil {
		return tablemodel.AccountBook{}, err
	}

	for statementType, parents := range input.Categories {
		for parentOrder, parent := range parents {
			parentCategory := tablemodel.Category{
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
			if err := tx.Table("categories").Omit("Budget", "SpecialType", "Frequent").Create(&parentCategory).Error; err != nil {
				return tablemodel.AccountBook{}, err
			}
			for childOrder, child := range parent.Childs {
				childCategory := tablemodel.Category{
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
				if err := tx.Table("categories").Omit("Budget", "SpecialType", "Frequent").Create(&childCategory).Error; err != nil {
					return tablemodel.AccountBook{}, err
				}
			}
		}
	}

	for _, asset := range input.Assets {
		assetType := asset.Type
		if assetType == "" {
			assetType = "deposit"
		}
		parentAsset := tablemodel.Asset{
			AccountBookID: created.ID,
			Name:          asset.Name,
			IconPath:      asset.IconPath,
			Type:          assetType,
			ParentID:      0,
			CreatorID:     input.UserID,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := tx.Table("assets").Omit("Order", "Frequent", "Remark").Create(&parentAsset).Error; err != nil {
			return tablemodel.AccountBook{}, err
		}
		for _, child := range asset.Childs {
			childAsset := tablemodel.Asset{
				AccountBookID: created.ID,
				Name:          child.Name,
				IconPath:      child.IconPath,
				Type:          assetType,
				ParentID:      parentAsset.ID,
				CreatorID:     input.UserID,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := tx.Table("assets").Omit("Order", "Frequent", "Remark").Create(&childAsset).Error; err != nil {
				return tablemodel.AccountBook{}, err
			}
		}
	}

	return created, nil
}
