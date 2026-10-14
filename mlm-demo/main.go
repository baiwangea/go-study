// go:build level

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

// ==================== 模型 ====================

type User struct {
	ID         string  `json:"id"`
	ParentID   string  `json:"parent_id"`
	Rank       string  `json:"rank"`
	Rate       float64 `json:"rate"`
	TeamVolume float64 `json:"team_volume"`
}

type RankConfig struct {
	Name           string  `json:"name"`
	Rate           float64 `json:"rate"`
	LeadershipRate float64 `json:"leadership_rate"`
	PeerRate       float64 `json:"peer_rate"`
	MaxGeneration  int     `json:"max_generation"`
}

type Order struct {
	ID           string  `json:"id"`
	BuyerID      string  `json:"buyer_id"`
	Amount       float64 `json:"amount"`
	IsRepurchase bool    `json:"is_repurchase"`
}

type BonusRecord struct {
	UserID    string
	BonusType string
	Amount    float64
	FromUser  string
	Level     int
	Remark    string
}

type Rules struct {
	ReferralRate  float64
	OverrideRates map[int]float64
	PointBonus    float64
	MatchingRate  float64
}

// ==================== 引擎 ====================

type Engine struct {
	users map[string]*User
	ranks map[string]*RankConfig
	rules *Rules
}

func NewEngine(users []*User, ranks []*RankConfig) *Engine {
	uMap := make(map[string]*User)
	for _, u := range users {
		uMap[u.ID] = u
	}
	rMap := make(map[string]*RankConfig)
	for _, r := range ranks {
		rMap[r.Name] = r
	}
	return &Engine{
		users: uMap,
		ranks: rMap,
		rules: &Rules{
			ReferralRate:  0.10,
			OverrideRates: map[int]float64{2: 0.02, 3: 0.01, 4: 0.005},
			PointBonus:    5,
			MatchingRate:  0.10,
		},
	}
}

func (e *Engine) ancestors(uid string) []*User {
	var list []*User
	cur := uid
	for {
		u, ok := e.users[cur]
		if !ok || u.ParentID == "" {
			break
		}
		p, ok := e.users[u.ParentID]
		if !ok {
			break
		}
		list = append(list, p)
		cur = p.ID
	}
	return list
}

func round(v float64) float64 { return math.Round(v*100) / 100 }

// ---------- 奖项计算 ----------

func (e *Engine) differential(o *Order) []BonusRecord {
	var res []BonusRecord
	buyer := e.users[o.BuyerID]
	if buyer == nil {
		return res
	}
	curRate := buyer.Rate
	for _, a := range e.ancestors(o.BuyerID) {
		diff := a.Rate - curRate
		if diff > 0 {
			res = append(res, BonusRecord{
				UserID: a.ID, BonusType: "级差奖", Amount: round(o.Amount * diff),
				FromUser: o.BuyerID, Remark: fmt.Sprintf("%.0f%%-%.0f%%", a.Rate*100, curRate*100),
			})
		}
		curRate = a.Rate
	}
	return res
}

func (e *Engine) referral(o *Order) []BonusRecord {
	b := e.users[o.BuyerID]
	if b == nil || b.ParentID == "" {
		return nil
	}
	return []BonusRecord{{
		UserID: b.ParentID, BonusType: "推荐奖", Amount: round(o.Amount * e.rules.ReferralRate),
		FromUser: o.BuyerID, Level: 1,
	}}
}

func (e *Engine) override(o *Order) []BonusRecord {
	var res []BonusRecord
	for i, a := range e.ancestors(o.BuyerID) {
		level := i + 1
		if level == 1 {
			continue
		}
		rate, ok := e.rules.OverrideRates[level]
		if !ok {
			continue
		}
		if rk, ok := e.ranks[a.Rank]; ok && level > rk.MaxGeneration {
			continue
		}
		res = append(res, BonusRecord{
			UserID: a.ID, BonusType: "越级奖", Amount: round(o.Amount * rate),
			FromUser: o.BuyerID, Level: level,
		})
	}
	return res
}

func (e *Engine) peer(o *Order) []BonusRecord {
	var res []BonusRecord
	buyer := e.users[o.BuyerID]
	if buyer == nil {
		return res
	}
	for _, a := range e.ancestors(o.BuyerID) {
		if a.Rank == buyer.Rank {
			if rk, ok := e.ranks[a.Rank]; ok && rk.PeerRate > 0 {
				res = append(res, BonusRecord{
					UserID: a.ID, BonusType: "平级奖", Amount: round(o.Amount * rk.PeerRate),
					FromUser: o.BuyerID, Remark: "同职级",
				})
			}
		}
	}
	return res
}

func (e *Engine) point(o *Order) []BonusRecord {
	var res []BonusRecord
	for i, a := range e.ancestors(o.BuyerID) {
		res = append(res, BonusRecord{
			UserID: a.ID, BonusType: "见点奖", Amount: e.rules.PointBonus,
			FromUser: o.BuyerID, Level: i + 1,
		})
	}
	return res
}

func (e *Engine) leadership() []BonusRecord {
	var res []BonusRecord
	for _, u := range e.users {
		rk, ok := e.ranks[u.Rank]
		if !ok || rk.LeadershipRate <= 0 || u.TeamVolume <= 0 {
			continue
		}
		res = append(res, BonusRecord{
			UserID: u.ID, BonusType: "团队管理奖",
			Amount: round(u.TeamVolume * rk.LeadershipRate),
			Remark: fmt.Sprintf("团队业绩 %.0f", u.TeamVolume),
		})
	}
	return res
}

// 统一处理一笔订单
func (e *Engine) ProcessOrder(o *Order) []BonusRecord {
	var all []BonusRecord
	all = append(all, e.differential(o)...)
	all = append(all, e.referral(o)...)
	all = append(all, e.override(o)...)
	all = append(all, e.peer(o)...)
	all = append(all, e.point(o)...)
	return all
}

// ==================== 工具 ====================

func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func main() {
	var users []*User
	var ranks []*RankConfig
	var order Order

	if err := loadJSON("data/users.json", &users); err != nil {
		panic(err)
	}
	if err := loadJSON("data/ranks.json", &ranks); err != nil {
		panic(err)
	}
	if err := loadJSON("data/order.json", &order); err != nil {
		panic(err)
	}

	engine := NewEngine(users, ranks)

	fmt.Println("===== 订单奖金（实时） =====")
	records := engine.ProcessOrder(&order)
	total := 0.0
	for _, r := range records {
		fmt.Printf("%-4s | %-10s | %8.2f 元 | 来自 %-4s | 代数 %d | %s\n",
			r.UserID, r.BonusType, r.Amount, r.FromUser, r.Level, r.Remark)
		total += r.Amount
	}
	fmt.Printf("订单总奖金: %.2f 元\n\n", total)

	fmt.Println("===== 月结团队管理奖 =====")
	for _, r := range engine.leadership() {
		fmt.Printf("%-4s | %-10s | %8.2f 元 | %s\n",
			r.UserID, r.BonusType, r.Amount, r.Remark)
	}
}
