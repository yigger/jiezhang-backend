package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/yigger/jiezhang-backend/internal/domain"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/sessioncache"
	"github.com/yigger/jiezhang-backend/internal/infrastructure/wechat"
	"github.com/yigger/jiezhang-backend/internal/repository"
)

var ErrLoginFailed = errors.New("login failed")

type CheckOpenIDService struct {
	users       repository.UserRepository
	wechat      wechat.Client
	tokenSecret string
	cache       sessioncache.Cache
}

func NewCheckOpenIDService(
	users repository.UserRepository,
	wechatClient wechat.Client,
	tokenSecret string,
	cache sessioncache.Cache,
) CheckOpenIDService {
	return CheckOpenIDService{
		users:       users,
		wechat:      wechatClient,
		tokenSecret: tokenSecret,
		cache:       cache,
	}
}

func (s CheckOpenIDService) Execute(ctx context.Context, code string) (string, error) {
	session, err := s.exchangeCodeWithRetry(ctx, strings.TrimSpace(code), 4)
	if err != nil {
		return "", ErrLoginFailed
	}

	user, err := s.users.FindByOpenID(ctx, session.OpenID)
	if err != nil {
		if !errors.Is(err, repository.ErrUserNotFound) {
			return "", err
		}
		user = domain.User{
			OpenID:     session.OpenID,
			SessionKey: session.SessionKey,
		}
		bookInput := repository.AccountBookCreateInput{
			Name:       "生活账簿",
			Categories: defaultCategories(),
			Assets:     defaultAssets(),
		}
		user, err = s.users.CreateWithInit(ctx, user, bookInput)
		if err != nil {
			return "", err
		}
	}

	cacheKey := user.RedisSessionKey()
	if cached, ok := s.cache.Get(cacheKey); ok && strings.TrimSpace(cached) != "" {
		return cached, nil
	}

	thirdSession, err := s.generateSecureToken(user.ID, session.SessionKey)
	if err != nil {
		return "", err
	}

	user.SessionKey = session.SessionKey
	user.ThirdSession = thirdSession
	if _, err := s.users.Save(ctx, user); err != nil {
		return "", err
	}

	s.cache.Set(cacheKey, thirdSession, 48*time.Hour)
	return thirdSession, nil
}

func (s CheckOpenIDService) exchangeCodeWithRetry(ctx context.Context, code string, maxRetries int) (wechat.SessionResponse, error) {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		resp, err := s.wechat.Code2Session(ctx, code)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if !isTimeout(err) {
			break
		}
	}

	return wechat.SessionResponse{}, lastErr
}

func (s CheckOpenIDService) generateSecureToken(userID int64, sessionKey string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	raw := fmt.Sprintf(
		"%d|%s|%d|%s|%s",
		userID,
		sessionKey,
		time.Now().Unix(),
		hex.EncodeToString(randomBytes),
		s.tokenSecret,
	)

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]), nil
}

func defaultCategories() map[string][]repository.AccountBookCategoryTemplate {
	return map[string][]repository.AccountBookCategoryTemplate{
		"expend": {
			{Name: "伙食餐饮", IconPath: "/images/category/035-meal.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "一日三餐", IconPath: "/images/category/004-diet.png"},
				{Name: "下午茶", IconPath: "/images/category/013-coffee.png"},
				{Name: "宵夜", IconPath: "/images/category/007-kebab.png"},
				{Name: "做饭食材", IconPath: "/images/category/004-diet.png"},
				{Name: "商场购物", IconPath: "/images/category/026-shopping-bag.png"},
				{Name: "水果", IconPath: "/images/category/fruit.png"},
				{Name: "茶水", IconPath: "/images/category/10.png"},
				{Name: "饮料", IconPath: "/images/category/061-water.png"},
				{Name: "烟酒", IconPath: "/images/category/060-pint.png"},
				{Name: "聚会消费", IconPath: "/images/category/010-beer-1.png"},
			}},
			{Name: "休闲娱乐", IconPath: "/images/category/032-game-pad.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "运动健身", IconPath: "/images/category/035-fitness.png"},
				{Name: "电影", IconPath: "/images/category/006-cinema.png"},
				{Name: "电子产品", IconPath: "/images/category/043-gaming.png"},
				{Name: "游戏", IconPath: "/images/category/041-gamepad.png"},
			}},
			{Name: "行车交通", IconPath: "/images/category/031-traffic-sign.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "公交", IconPath: "/images/category/038-school-bus.png"},
				{Name: "地铁", IconPath: "/images/category/a.png"},
				{Name: "打车租车", IconPath: "/images/category/chuzuche.png"},
			}},
			{Name: "美容护肤", IconPath: "/images/category/028-fashion.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "化妆品", IconPath: "/images/category/makeup.png"},
				{Name: "面膜", IconPath: "/images/category/054-mask.png"},
				{Name: "口红", IconPath: "/images/category/056-lipstick.png"},
				{Name: "香水", IconPath: "/images/category/perfume.png"},
				{Name: "指甲", IconPath: "/images/category/053-cologne.png"},
			}},
			{Name: "衣服饰品", IconPath: "/images/category/018-tshirt.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "衣服", IconPath: "/images/category/031-jacket.png"},
				{Name: "裤子", IconPath: "/images/category/029-jeans.png"},
				{Name: "鞋子", IconPath: "/images/category/027-jogging.png"},
				{Name: "包包", IconPath: "/images/category/handbag.png"},
			}},
			{Name: "交流通讯", IconPath: "/images/category/011-smartphone.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "手机费", IconPath: "/images/category/037-smartphone.png"},
				{Name: "上网费", IconPath: "/images/category/036-laptop.png"},
			}},
			{Name: "学习进修", IconPath: "/images/category/007-laptop.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "书籍费用", IconPath: "/images/category/033-open-book.png"},
				{Name: "培训学习", IconPath: "/images/category/diary.png"},
			}},
			{Name: "居家物业", IconPath: "/images/category/001-house.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "房租", IconPath: "/images/category/hee.png"},
				{Name: "日常用品", IconPath: "/images/category/towel.png"},
			}},
			{Name: "医疗保障", IconPath: "/images/category/022-hospital-1.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "药品", IconPath: "/images/category/jiaonang.png"},
				{Name: "就诊", IconPath: "/images/category/021-emergency-kit.png"},
				{Name: "保健", IconPath: "/images/category/024-hospital.png"},
			}},
		},
		"income": {
			{Name: "职业收入", IconPath: "/images/category/cash.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "工资收入", IconPath: "/images/category/cash.png"},
				{Name: "利息收入", IconPath: "/images/category/bank.png"},
				{Name: "加班收入", IconPath: "/images/category/jiaban.png"},
				{Name: "投资收入", IconPath: "/images/category/money-bag.png"},
			}},
			{Name: "其它收入", IconPath: "/images/category/caipiao.png", Childs: []repository.AccountBookCategoryChildTemplate{
				{Name: "中奖收入", IconPath: "/images/category/caipiao.png"},
				{Name: "意外来钱", IconPath: "/images/category/chuangyi.png"},
				{Name: "红包收入", IconPath: "/images/category/hongbao.png"},
			}},
		},
	}
}

func defaultAssets() []repository.AccountBookAssetTemplate {
	return []repository.AccountBookAssetTemplate{
		{Name: "现金账户", IconPath: "/images/asset/coin-2.png", Type: "deposit", Childs: []repository.AccountBookAssetChildTemplate{
			{Name: "现金", IconPath: "/images/asset/wallet-1.png"},
			{Name: "银行卡", IconPath: "/images/asset/credit-card-6.png"},
		}},
		{Name: "虚拟账户", IconPath: "/images/asset/coin-4.png", Type: "deposit", Childs: []repository.AccountBookAssetChildTemplate{
			{Name: "支付宝", IconPath: "/images/asset/12.png"},
			{Name: "微信钱包", IconPath: "/images/asset/9.png"},
		}},
		{Name: "负债账户", IconPath: "/images/asset/justice-scale.png", Type: "debt", Childs: []repository.AccountBookAssetChildTemplate{
			{Name: "蚂蚁花呗", IconPath: "/images/asset/huabei.png"},
			{Name: "京东白条", IconPath: "/images/asset/jdbaitiao.png"},
			{Name: "信用卡", IconPath: "/images/asset/credit-card-1.png"},
		}},
		{Name: "投资账户", IconPath: "/images/asset/graph.png", Type: "deposit", Childs: []repository.AccountBookAssetChildTemplate{
			{Name: "基金账户", IconPath: "/images/asset/1.png"},
			{Name: "余额宝", IconPath: "/images/asset/yuebao.png"},
			{Name: "股票账户", IconPath: "/images/asset/graph-3.png"},
		}},
	}
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	return errors.Is(err, context.DeadlineExceeded)
}
