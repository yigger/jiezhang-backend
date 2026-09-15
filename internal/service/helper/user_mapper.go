package helper

import (
	tablemodel "github.com/yigger/jiezhang-backend/internal/model"
	"github.com/yigger/jiezhang-backend/internal/types"
)

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func ToUserContext(model tablemodel.User) types.UserContext {
	return types.UserContext{
		ID:               model.ID,
		UID:              model.UID,
		Nickname:         ptrStr(model.Nickname),
		Email:            ptrStr(model.Email),
		OpenID:           model.OpenID,
		SessionKey:       ptrStr(model.SessionKey),
		ThirdSession:     ptrStr(model.ThirdSession),
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		AccountBookId:    model.AccountBookID,
		ThemeID:          model.ThemeID,
		AvatarUrl:        ptrStr(model.AvatarURL),
		Country:          ptrStr(model.Country),
		City:             ptrStr(model.City),
		Gender:           model.Gender,
		Language:         ptrStr(model.Language),
		Province:         ptrStr(model.Province),
		BGAvatarID:       model.BGAvatarID,
		BGAvatarURL:      ptrStr(model.BGAvatarURL),
		Remind:           model.Remind,
		HiddenAssetMoney: model.HiddenAssetMoney,
		AlreadyLogin:     model.AlreadyLogin,
	}
}

func UserModel(user types.UserContext) tablemodel.User {
	return tablemodel.User{
		ID:               user.ID,
		UID:              user.UID,
		Nickname:         strPtr(user.Nickname),
		Email:            strPtr(user.Email),
		OpenID:           user.OpenID,
		SessionKey:       strPtr(user.SessionKey),
		ThirdSession:     strPtr(user.ThirdSession),
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		AccountBookID:    user.AccountBookId,
		ThemeID:          user.ThemeID,
		AvatarURL:        strPtr(user.AvatarUrl),
		Country:          strPtr(user.Country),
		City:             strPtr(user.City),
		Gender:           user.Gender,
		Language:         strPtr(user.Language),
		Province:         strPtr(user.Province),
		BGAvatarID:       user.BGAvatarID,
		BGAvatarURL:      strPtr(user.BGAvatarURL),
		Remind:           user.Remind,
		HiddenAssetMoney: user.HiddenAssetMoney,
		AlreadyLogin:     user.AlreadyLogin,
	}
}
