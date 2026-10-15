//go:build level

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
)

/*
════════════════════════════════════════════════════════════════════════════
MLM 分销奖励机制演示系统
════════════════════════════════════════════════════════════════════════════

本文件演示了多层级营销（MLM）中的12种常见奖励算法：
1. 级差奖 (Differential)      - 基于职级差异
2. 推荐奖 (Referral)          - 直接推荐奖励
3. 越级奖 (Override)          - 多代奖励
4. 平级奖 (Peer)              - 同职级激励
5. 团队管理奖 (Leadership)     - 团队整体奖励
6. 见点奖 (Point)             - 每单固定奖励
7. 复购奖 (Repurchase)        - 老客户复购
8. 育成奖 (Mentoring)         - 培养人才
9. 晋级奖 (Advancement)       - 职级提升
10. 感恩奖 (Matching)         - 分享下级奖金
11. 层奖 (Layer)              - 按层级奖励
12. 全局分红 (Dividend)       - 公司业绩分红

════════════════════════════════════════════════════════════════════════════
*/

// ════════════════════════════════════════════════════════════════════════════
// 数据模型
// ════════════════════════════════════════════════════════════════════════════

// User 用户模型 - 表示分销网络中的一个节点
type User struct {
	ID         string  `json:"id"`          // 用户唯一ID
	Name       string  `json:"name"`        // 用户名称
	ParentID   string  `json:"parent_id"`   // 推荐人ID（上级）
	Rank       string  `json:"rank"`        // 当前职级名称
	Rate       float64 `json:"rate"`        // 职级对应的基础比例
	TeamVolume float64 `json:"team_volume"` // 团队总业绩（用于团队管理奖）
}

// RankConfig 职级配置 - 定义每个职级的权益
type RankConfig struct {
	Name           string  `json:"name"`            // 职级名称（如：VIP、经理、总监）
	Level          int     `json:"level"`           // 职级等级（数字越大越高）
	Rate           float64 `json:"rate"`            // 基础分润比例（用于级差奖）
	LeadershipRate float64 `json:"leadership_rate"` // 团队管理奖比例
	PeerRate       float64 `json:"peer_rate"`       // 平级奖比例
	MaxGeneration  int     `json:"max_generation"`  // 最大拿奖代数（越级奖用）
}

// Order 订单模型 - 触发奖励计算的核心数据
type Order struct {
	ID           string  `json:"id"`            // 订单ID
	BuyerID      string  `json:"buyer_id"`      // 下单用户ID
	Amount       float64 `json:"amount"`        // 订单金额
	IsRepurchase bool    `json:"is_repurchase"` // 是否复购订单
	ProductType  string  `json:"product_type"`  // 产品类型（用于差异化奖励）
}

// BonusRecord 奖金记录 - 单笔奖励的详细信息
type BonusRecord struct {
	UserID    string  // 获得奖金的用户ID
	UserName  string  // 获得奖金的用户名称
	BonusType string  // 奖金类型（如：级差奖、推荐奖等）
	Amount    float64 // 奖金金额
	FromUser  string  // 来源用户ID（触发奖励的人）
	FromName  string  // 来源用户名称
	Level     int     // 代数（第几代，0表示不依赖代数）
	Remark    string  // 备注说明
}

// Rules 全局奖励规则配置
type Rules struct {
	// 推荐奖：直接推荐人拿的固定比例
	// 示例：10% 表示下单1000元，推荐人拿100元
	ReferralRate float64

	// 复购奖：老客户复购时推荐人的额外奖励比例
	// 通常高于首单推荐奖，鼓励维护老客户
	RepurchaseRate float64

	// 越级奖：按代数递减的奖励比例
	// map[代数]比例，如 map[2]0.05 表示第2代拿5%
	OverrideRates map[int]float64

	// 见点奖：每产生一单，上级拿的固定金额
	// 示例：5元/单，不管订单金额多少
	PointBonus float64

	// 感恩奖/匹配奖：下级拿到奖金后，上级再匹配的比例
	// 示例：20% 表示下级拿100元奖金，上级再拿20元匹配奖
	MatchingRate float64

	// 层奖：按层级给予的固定奖励
	// map[层数]金额，常见于矩阵制
	LayerBonus map[int]float64

	// 育成奖：下级晋升到指定职级时给上级的一次性奖励
	// map[职级名称]奖金金额
	MentoringBonus map[string]float64

	// 晋升奖：用户自己晋升时获得的一次性奖励
	// map[新职级名称]奖金金额
	AdvancementBonus map[string]float64

	// 全局分红池比例：从公司总业绩中提取的比例
	// 示例：5% 表示月业绩1000万，分红池50万
	DividendPoolRate float64
}

// ════════════════════════════════════════════════════════════════════════════
// 奖励计算引擎
// ════════════════════════════════════════════════════════════════════════════

type Engine struct {
	users       map[string]*User       // 所有用户的映射表 [ID -> User]
	ranks       map[string]*RankConfig // 所有职级的配置 [名称 -> Config]
	rules       *Rules                 // 全局奖励规则
	totalVolume float64                // 公司总业绩（用于全局分红）
	bonusPool   map[string]float64     // 临时奖金池，记录每个人的累计奖金
}

func NewEngine(users []*User, ranks []*RankConfig) *Engine {
	// 构建用户映射表，方便快速查找
	uMap := make(map[string]*User)
	for _, u := range users {
		uMap[u.ID] = u
	}

	// 构建职级映射表
	rMap := make(map[string]*RankConfig)
	for _, r := range ranks {
		rMap[r.Name] = r
	}

	return &Engine{
		users:     uMap,
		ranks:     rMap,
		bonusPool: make(map[string]float64),
		rules: &Rules{
			ReferralRate:   0.10, // 推荐奖10%
			RepurchaseRate: 0.15, // 复购奖15%（比首单高）
			OverrideRates: map[int]float64{
				2: 0.05, // 第2代拿5%
				3: 0.03, // 第3代拿3%
				4: 0.02, // 第4代拿2%
				5: 0.01, // 第5代拿1%
			},
			PointBonus:   5.0,  // 每单固定5元
			MatchingRate: 0.20, // 匹配20%
			LayerBonus: map[int]float64{
				1: 50, // 第1层每人50元
				2: 30, // 第2层每人30元
				3: 20, // 第3层每人20元
			},
			MentoringBonus: map[string]float64{
				"经理":  1000,  // 下级升到经理，给1000元
				"总监":  5000,  // 下级升到总监，给5000元
				"合伙人": 20000, // 下级升到合伙人，给20000元
			},
			AdvancementBonus: map[string]float64{
				"经理":  1000,  // 自己升到经理，拿1000元
				"总监":  5000,  // 自己升到总监，拿5000元
				"合伙人": 20000, // 自己升到合伙人，拿20000元
			},
			DividendPoolRate: 0.05, // 全局分红池占总业绩5%
		},
	}
}

// ancestors 获取某个用户的所有上级（从直接推荐人到最顶层）
// 返回值按代数排序：[第1代(直接推荐人), 第2代, 第3代, ...]
func (e *Engine) ancestors(uid string) []*User {
	var list []*User
	cur := uid

	// 从当前用户开始，沿着 ParentID 向上爬
	for {
		u, ok := e.users[cur]
		if !ok || u.ParentID == "" {
			break // 到达顶层或找不到用户
		}

		parent, ok := e.users[u.ParentID]
		if !ok {
			break // 上级不存在
		}

		list = append(list, parent)
		cur = parent.ID
	}

	return list
}

// round 金额四舍五入到分（保留2位小数）
func round(v float64) float64 {
	return math.Round(v*100) / 100
}

// ════════════════════════════════════════════════════════════════════════════
// 奖励算法实现
// ════════════════════════════════════════════════════════════════════════════

// ════════════════════════════════════════════════════════════════════════════
// 1. 级差奖（Differential Bonus）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 职级高的人从职级低的人的订单中拿"职级差"对应的比例。
// 这种设计鼓励用户提升职级，因为职级越高，能拿的比例越大。
//
// 【计算公式】
// 级差奖 = 订单金额 × (自己的职级比例 - 下单人的职级比例)
//
// 【示例】
// 假设职级比例：
//   - 普通会员: 5%
//   - VIP: 10%
//   - 经理: 15%
//   - 总监: 20%
//
// 层级关系：张三(总监20%) → 李四(经理15%) → 王五(VIP10%) → 赵六(普通5%)
//
// 当赵六下单1000元时：
//   - 王五拿：1000 × (10% - 5%) = 50元   （自己10% - 赵六5%）
//   - 李四拿：1000 × (15% - 10%) = 50元  （自己15% - 王五10%）
//   - 张三拿：1000 × (20% - 15%) = 50元  （自己20% - 李四15%）
//
// 【关键点】
// 1. ❌ 不依赖代数 - 不管第几代，只看职级差
// 2. ✅ 强依赖职级 - 职级相同或更低的人拿不到
// 3. 💰 逐级递减 - 每一层只拿"增量"部分，不会重复
// 4. 🎯 激励晋升 - 想拿更多，就得提升职级
func (e *Engine) differential(o *Order) []BonusRecord {
	var records []BonusRecord

	// 1. 获取下单人信息
	buyer := e.users[o.BuyerID]
	if buyer == nil {
		return records
	}

	// 2. 从下单人的职级比例开始
	currentRate := buyer.Rate

	// 3. 向上遍历所有上级
	for _, ancestor := range e.ancestors(o.BuyerID) {
		// 4. 计算职级差
		diff := ancestor.Rate - currentRate

		// 5. 只有职级更高的人才能拿奖
		if diff > 0 {
			records = append(records, BonusRecord{
				UserID:    ancestor.ID,
				UserName:  ancestor.Name,
				BonusType: "级差奖",
				Amount:    round(o.Amount * diff),
				FromUser:  o.BuyerID,
				FromName:  buyer.Name,
				Remark:    fmt.Sprintf("%.0f%% - %.0f%% = %.0f%%", ancestor.Rate*100, currentRate*100, diff*100),
			})
		}

		// 6. 更新当前比例，继续向上
		// 这样确保下一级只拿"增量"，不会重复计算
		currentRate = ancestor.Rate
	}

	return records
}

// ════════════════════════════════════════════════════════════════════════════
// 2. 推荐奖/直推奖（Referral / Sponsor Bonus）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 最简单直接的奖励：直接推荐人拿固定比例。
// 这是MLM系统中最基础、最常见的奖励机制。
//
// 【计算公式】
// 推荐奖 = 订单金额 × 推荐奖比例
//
// 【示例】
// 张三直接推荐了李四（李四的 parent_id = 张三的 id）
// 李四下单1000元，推荐奖比例10%
// 张三拿：1000 × 10% = 100元
//
// 【关键点】
// 1. ✅ 仅限第1代 - 只有直接推荐人能拿
// 2. 📊 可选职级影响 - 可以根据职级设置不同比例
// 3. 💵 固定比例 - 计算简单，容易理解
// 4. 🎁 即时激励 - 推荐成功就有收益
//
// 【变种】
// - 有些系统根据推荐人职级给不同比例
// - 有些系统首单和复购给不同比例（见复购奖）
func (e *Engine) referral(o *Order) []BonusRecord {
	// 1. 获取下单人信息
	buyer := e.users[o.BuyerID]
	if buyer == nil || buyer.ParentID == "" {
		return nil // 没有推荐人，无推荐奖
	}

	// 2. 获取推荐人信息
	sponsor := e.users[buyer.ParentID]
	if sponsor == nil {
		return nil
	}

	// 3. 根据是否复购选择比例
	rate := e.rules.ReferralRate
	bonusType := "推荐奖"
	if o.IsRepurchase {
		rate = e.rules.RepurchaseRate
		bonusType = "复购奖"
	}

	// 4. 计算并返回奖金
	return []BonusRecord{{
		UserID:    sponsor.ID,
		UserName:  sponsor.Name,
		BonusType: bonusType,
		Amount:    round(o.Amount * rate),
		FromUser:  o.BuyerID,
		FromName:  buyer.Name,
		Level:     1, // 第1代（直推）
		Remark:    fmt.Sprintf("%.0f%%", rate*100),
	}}
}

// ════════════════════════════════════════════════════════════════════════════
// 3. 越级奖/代数奖（Override / Generation Bonus）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 从第2代开始，每一代都拿固定比例，通常逐代递减。
// 鼓励发展深度团队，而不仅仅是直推。
//
// 【计算公式】
// 越级奖 = 订单金额 × 对应代数的比例
//
// 【示例】
// 层级关系：张三 → 李四 → 王五 → 赵六 → 钱七
//
// 越级奖比例配置：
//
//	第2代: 5%
//	第3代: 3%
//	第4代: 2%
//	第5代: 1%
//
// 当钱七下单1000元时：
//
//	李四拿不到（他只能拿推荐奖，王五是他的第1代）
//	王五拿：1000 × 5% = 50元  （钱七是他的第2代）
//	李四拿：1000 × 3% = 30元  （钱七是他的第3代）
//	张三拿：1000 × 2% = 20元  （钱七是他的第4代）
//
// 【关键点】
// 1. ✅ 依赖代数 - 从第2代开始算
// 2. 📉 逐代递减 - 越远拿得越少
// 3. 🎯 激励深度 - 鼓励发展多层团队
// 4. 🔒 职级限制 - 高职级才能拿更多代数
//
// 【注意】
// 通常会配合职级设置最大代数限制：
// - VIP: 只能拿到第3代
// - 经理: 可以拿到第5代
// - 总监: 可以拿到第10代
func (e *Engine) override(o *Order) []BonusRecord {
	var records []BonusRecord

	// 1. 获取下单人信息
	buyer := e.users[o.BuyerID]
	if buyer == nil {
		return records
	}

	// 2. 遍历所有上级
	for i, ancestor := range e.ancestors(o.BuyerID) {
		level := i + 1 // 第几代（1-based）

		// 3. 跳过第1代（第1代走推荐奖）
		if level == 1 {
			continue
		}

		// 4. 检查这一代是否有配置的越级奖比例
		rate, ok := e.rules.OverrideRates[level]
		if !ok {
			continue // 这一代没有配置，跳过
		}

		// 5. 检查职级限制（高职级才能拿更远的代数）
		if rankConfig, ok := e.ranks[ancestor.Rank]; ok {
			if level > rankConfig.MaxGeneration {
				continue // 超过了该职级的最大代数限制
			}
		}

		// 6. 计算并记录奖金
		records = append(records, BonusRecord{
			UserID:    ancestor.ID,
			UserName:  ancestor.Name,
			BonusType: "越级奖",
			Amount:    round(o.Amount * rate),
			FromUser:  o.BuyerID,
			FromName:  buyer.Name,
			Level:     level,
			Remark:    fmt.Sprintf("第%d代 %.0f%%", level, rate*100),
		})
	}

	return records
}

// ════════════════════════════════════════════════════════════════════════════
// 4. 平级奖（Peer / Same Rank Bonus）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 当下级达到与自己相同的职级时，上级额外获得奖励。
// 防止高职级"躺赚"，鼓励持续辅导下级成长。
//
// 【计算公式】
// 平级奖 = 订单金额 × 平级奖比例
//
// 【示例】
// 张三：总监
// 李四：总监（张三的下级）
// 王五：经理（李四的下级）
//
// 当王五下单1000元时：
//
//	李四拿：正常的级差奖、推荐奖等
//	张三拿：1000 × 5% = 50元（因为李四也是总监，触发平级奖）
//
// 【关键点】
// 1. ❌ 不依赖代数 - 只要下级链条中有平级就触发
// 2. ✅ 强依赖职级 - 必须职级相同
// 3. 🎯 防止压级 - 鼓励下级晋升，而不是故意压着
// 4. 💪 持续激励 - 下级成长了，上级也受益
//
// 【注意】
// 平级奖通常只给直接的"平级"关系，而不是所有相同职级的人。
func (e *Engine) peer(o *Order) []BonusRecord {
	var records []BonusRecord

	// 1. 获取下单人信息
	buyer := e.users[o.BuyerID]
	if buyer == nil {
		return records
	}

	// 2. 遍历所有上级
	for _, ancestor := range e.ancestors(o.BuyerID) {
		// 3. 检查是否与下单人职级相同
		if ancestor.Rank == buyer.Rank {
			// 4. 获取该职级的平级奖配置
			if rankConfig, ok := e.ranks[ancestor.Rank]; ok && rankConfig.PeerRate > 0 {
				records = append(records, BonusRecord{
					UserID:    ancestor.ID,
					UserName:  ancestor.Name,
					BonusType: "平级奖",
					Amount:    round(o.Amount * rankConfig.PeerRate),
					FromUser:  o.BuyerID,
					FromName:  buyer.Name,
					Remark:    fmt.Sprintf("同为%s职级", ancestor.Rank),
				})
			}
		}
	}

	return records
}

// ════════════════════════════════════════════════════════════════════════════
// 5. 见点奖/见单奖（Point / Order Bonus）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 每产生一单，上级拿固定金额，与订单金额无关。
// 适合高频低价产品，快速激励。
//
// 【计算公式】
// 见点奖 = 固定金额
//
// 【示例】
// 配置：每单固定奖励5元，最多拿5代
//
// 层级关系：张三 → 李四 → 王五 → 赵六
//
// 当赵六下单（不管金额多少）时：
//
//	王五拿：5元（第1代）
//	李四拿：5元（第2代）
//	张三拿：5元（第3代）
//
// 【关键点】
// 1. ✅ 依赖代数 - 通常限制在3-5代
// 2. ❌ 不依赖职级 - 大家拿的都一样
// 3. 💰 固定金额 - 计算简单，订单多少都是固定数
// 4. ⚡ 快速激励 - 每单都有，积少成多
//
// 【适用场景】
// - 快消品（如日用品、化妆品）
// - 低价产品（单价几十元的）
// - 高频交易（每天都有订单的）
func (e *Engine) point(o *Order) []BonusRecord {
	var records []BonusRecord

	// 1. 获取下单人信息
	buyer := e.users[o.BuyerID]
	if buyer == nil {
		return records
	}

	// 2. 遍历所有上级，每人拿固定金额
	for i, ancestor := range e.ancestors(o.BuyerID) {
		level := i + 1

		// 3. 可以设置最大代数限制（这里示例为不限）
		// if level > 5 { break } // 例如：只拿5代

		records = append(records, BonusRecord{
			UserID:    ancestor.ID,
			UserName:  ancestor.Name,
			BonusType: "见点奖",
			Amount:    e.rules.PointBonus,
			FromUser:  o.BuyerID,
			FromName:  buyer.Name,
			Level:     level,
			Remark:    fmt.Sprintf("第%d代固定%.0f元", level, e.rules.PointBonus),
		})
	}

	return records
}

// ════════════════════════════════════════════════════════════════════════════
// 6. 团队管理奖/领导奖（Leadership / Management Bonus）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 根据整个团队的业绩，按照职级给予管理奖励。
// 通常是月结或周结，不是实时计算。
//
// 【计算公式】
// 团队管理奖 = 团队总业绩 × 职级管理比例
//
// 【示例】
// 张三：总监，管理比例3%
// 张三的整个团队本月业绩：50万元
//
// 张三本月团队管理奖 = 50万 × 3% = 15000元
//
// 【职级管理比例示例】
// - 普通会员：0%
// - VIP：1%
// - 经理：2%
// - 总监：3%
// - 合伙人：5%
//
// 【关键点】
// 1. ❌ 不依赖代数 - 统计整个团队，不分代数
// 2. ✅ 强依赖职级 - 职级越高，比例越大
// 3. 📊 周期结算 - 通常月结，不是实时
// 4. 🎯 鼓励管理 - 团队大了，收益就大
//
// 【注意】
// 团队业绩通常不包括自己的业绩，只算下级的。
// 有些系统会排除平级及以上的下级业绩，防止重复计算。
func (e *Engine) leadership() []BonusRecord {
	var records []BonusRecord

	// 遍历所有用户
	for _, user := range e.users {
		// 1. 获取该用户的职级配置
		rankConfig, ok := e.ranks[user.Rank]
		if !ok {
			continue
		}

		// 2. 检查是否有管理奖比例
		if rankConfig.LeadershipRate <= 0 {
			continue
		}

		// 3. 检查团队业绩（这个值应该是预先统计好的）
		if user.TeamVolume <= 0 {
			continue
		}

		// 4. 计算团队管理奖
		records = append(records, BonusRecord{
			UserID:    user.ID,
			UserName:  user.Name,
			BonusType: "团队管理奖",
			Amount:    round(user.TeamVolume * rankConfig.LeadershipRate),
			Remark:    fmt.Sprintf("团队业绩%.0f × %.1f%%", user.TeamVolume, rankConfig.LeadershipRate*100),
		})
	}

	return records
}

// ════════════════════════════════════════════════════════════════════════════
// 7. 感恩奖/匹配奖（Matching / Gratitude Bonus）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 下级拿到的奖金，上级再匹配一定比例。
// 让上级分享下级的成功，激励辅导。
//
// 【计算公式】
// 感恩奖 = 下级获得的奖金总和 × 匹配比例
//
// 【示例】
// 李四本月获得各类奖金共：10000元
// 张三是李四的上级，匹配比例20%
//
// 张三感恩奖 = 10000 × 20% = 2000元
//
// 【关键点】
// 1. ✅ 依赖代数 - 通常限制在1-3代
// 2. 可选职级影响 - 高职级匹配比例更高
// 3. 💡 "奖金的奖金" - 二次分配
// 4. 🤝 激励辅导 - 下级赚得多，上级也多
//
// 【注意】
// 感恩奖通常在所有其他奖金计算完后再计算，避免循环引用。
// 有些系统只匹配特定类型的奖金（如推荐奖、越级奖），不匹配全部。
func (e *Engine) matching() []BonusRecord {
	var records []BonusRecord

	// 遍历所有用户
	for _, user := range e.users {
		// 获取该用户本期累计获得的奖金
		totalBonus := e.bonusPool[user.ID]
		if totalBonus <= 0 {
			continue
		}

		// 获取该用户的上级
		if user.ParentID == "" {
			continue
		}
		sponsor := e.users[user.ParentID]
		if sponsor == nil {
			continue
		}

		// 计算匹配奖
		matchingAmount := round(totalBonus * e.rules.MatchingRate)
		records = append(records, BonusRecord{
			UserID:    sponsor.ID,
			UserName:  sponsor.Name,
			BonusType: "感恩奖",
			Amount:    matchingAmount,
			FromUser:  user.ID,
			FromName:  user.Name,
			Remark:    fmt.Sprintf("下级奖金%.2f × %.0f%%", totalBonus, e.rules.MatchingRate*100),
		})
	}

	return records
}

// ════════════════════════════════════════════════════════════════════════════
// 8. 全局分红（Global Pool / Dividend）
// ════════════════════════════════════════════════════════════════════════════
//
// 【核心思想】
// 从公司总业绩中提取一定比例作为分红池，按职级权重分配。
// 让高层分享公司整体成长，属于长期激励。
//
// 【计算步骤】
// 1. 计算公司总业绩
// 2. 提取分红池 = 总业绩 × 分红池比例
// 3. 统计有资格参与分红的人数和权重
// 4. 按权重分配分红池
//
// 【示例】
// 公司月业绩：1000万
// 分红池比例：5%
// 分红池总额：1000万 × 5% = 50万
//
// 职级权重配置：
// - 经理：1份
// - 总监：3份
// - 合伙人：10份
//
// 假设有：
// - 10个经理（共10份）
// - 5个总监（共15份）
// - 2个合伙人（共20份）
// 总权重 = 10 + 15 + 20 = 45份
//
// 每份价值 = 50万 / 45 = 11111元
//
// 每个经理分得：11111元
// 每个总监分得：11111 × 3 = 33333元
// 每个合伙人分得：11111 × 10 = 111110元
//
// 【关键点】
// 1. ❌ 不依赖代数 - 全局性奖励
// 2. ✅ 强依赖职级 - 职级越高，权重越大
// 3. 📊 周期结算 - 通常月结或季结
// 4. 🌍 共享成长 - 公司好，大家都好
func (e *Engine) dividend() []BonusRecord {
	var records []BonusRecord

	// 1. 计算分红池总额
	dividendPool := e.totalVolume * e.rules.DividendPoolRate
	if dividendPool <= 0 {
		return records
	}

	// 2. 定义职级权重（这里简化处理，实际应该在配置中）
	rankWeights := map[string]int{
		"普通会员": 0,  // 普通会员不参与分红
		"VIP":  0,  // VIP不参与分红
		"经理":   1,  // 经理1份
		"总监":   3,  // 总监3份
		"合伙人":  10, // 合伙人10份
	}

	// 3. 统计总权重和每个职级的人数
	totalWeight := 0
	rankCount := make(map[string]int)
	for _, user := range e.users {
		if weight, ok := rankWeights[user.Rank]; ok && weight > 0 {
			totalWeight += weight
			rankCount[user.Rank]++
		}
	}

	if totalWeight == 0 {
		return records // 没有人有资格分红
	}

	// 4. 计算每份的价值
	perShare := dividendPool / float64(totalWeight)

	// 5. 给每个有资格的人分配
	for _, user := range e.users {
		weight, ok := rankWeights[user.Rank]
		if !ok || weight == 0 {
			continue
		}

		amount := round(perShare * float64(weight))
		records = append(records, BonusRecord{
			UserID:    user.ID,
			UserName:  user.Name,
			BonusType: "全局分红",
			Amount:    amount,
			Remark:    fmt.Sprintf("%s职级%d份权重", user.Rank, weight),
		})
	}

	return records
}

// ════════════════════════════════════════════════════════════════════════════
// 主要业务方法
// ════════════════════════════════════════════════════════════════════════════

// ProcessOrder 处理单笔订单，计算所有实时奖励
// 实时奖励包括：级差奖、推荐奖、越级奖、平级奖、见点奖
func (e *Engine) ProcessOrder(o *Order) []BonusRecord {
	var all []BonusRecord

	// 累加所有奖励
	all = append(all, e.differential(o)...)
	all = append(all, e.referral(o)...)
	all = append(all, e.override(o)...)
	all = append(all, e.peer(o)...)
	all = append(all, e.point(o)...)

	// 累计订单金额到总业绩（用于全局分红）
	e.totalVolume += o.Amount

	// 将奖金累加到奖金池（用于感恩奖）
	for _, record := range all {
		e.bonusPool[record.UserID] += record.Amount
	}

	return all
}

// ProcessPeriodEnd 处理周期结束（月结/周结）
// 周期性奖励包括：团队管理奖、感恩奖、全局分红
func (e *Engine) ProcessPeriodEnd() []BonusRecord {
	var all []BonusRecord

	// 计算周期性奖励
	all = append(all, e.leadership()...)
	all = append(all, e.matching()...)
	all = append(all, e.dividend()...)

	return all
}

// ════════════════════════════════════════════════════════════════════════════
// 工具函数
// ════════════════════════════════════════════════════════════════════════════

// loadJSON 从文件加载JSON数据
func loadJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取文件失败 %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("解析JSON失败 %s: %v", path, err)
	}
	return nil
}

// ════════════════════════════════════════════════════════════════════════════
// 主程序
// ════════════════════════════════════════════════════════════════════════════

func main() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🎁 MLM 分销奖励机制演示系统")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// ─────────────────────────────────────────────────────────────────────────
	// 1. 加载数据
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("📂 加载数据...")

	var users []*User
	var ranks []*RankConfig
	var order Order

	if err := loadJSON("data/users.json", &users); err != nil {
		fmt.Printf("❌ 加载用户数据失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 用户数据: %d 人\n", len(users))

	if err := loadJSON("data/ranks.json", &ranks); err != nil {
		fmt.Printf("❌ 加载职级数据失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 职级配置: %d 个职级\n", len(ranks))

	if err := loadJSON("data/order.json", &order); err != nil {
		fmt.Printf("❌ 加载订单数据失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 订单数据: 订单金额 %.2f 元\n", order.Amount)
	fmt.Println()

	// ─────────────────────────────────────────────────────────────────────────
	// 2. 初始化引擎
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println("⚙️  初始化奖励引擎...")
	engine := NewEngine(users, ranks)
	fmt.Println("   ✅ 引擎初始化完成")
	fmt.Println()

	// ─────────────────────────────────────────────────────────────────────────
	// 3. 处理订单，计算实时奖励
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("💰 实时奖励结算（订单触发）")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("订单信息：%s 购买 %.2f 元\n", order.BuyerID, order.Amount)
	fmt.Println(strings.Repeat("-", 80))

	records := engine.ProcessOrder(&order)

	// 按奖励类型分组统计
	typeStats := make(map[string]float64)
	totalBonus := 0.0

	// 打印详细记录
	fmt.Printf("%-6s | %-12s | %10s | %-10s | 代数 | %-20s\n",
		"用户ID", "奖励类型", "金额(元)", "来源用户", "备注")
	fmt.Println(strings.Repeat("-", 80))

	for _, r := range records {
		fmt.Printf("%-6s | %-12s | %10.2f | %-10s | %4d | %-20s\n",
			r.UserID, r.BonusType, r.Amount, r.FromUser, r.Level, r.Remark)

		typeStats[r.BonusType] += r.Amount
		totalBonus += r.Amount
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("订单总奖金: %.2f 元 (订单金额的 %.1f%%)\n",
		totalBonus, totalBonus/order.Amount*100)
	fmt.Println()

	// ─────────────────────────────────────────────────────────────────────────
	// 4. 奖励类型汇总
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("📊 奖励类型汇总")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-15s | %10s | %8s\n", "奖励类型", "总金额(元)", "占比")
	fmt.Println(strings.Repeat("-", 80))

	for bonusType, amount := range typeStats {
		percentage := amount / totalBonus * 100
		fmt.Printf("%-15s | %10.2f | %7.1f%%\n", bonusType, amount, percentage)
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-15s | %10.2f | %7.1f%%\n", "合计", totalBonus, 100.0)
	fmt.Println()

	// ─────────────────────────────────────────────────────────────────────────
	// 5. 周期性奖励（月结）
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("📅 周期性奖励结算（月结）")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("包含：团队管理奖、感恩奖、全局分红")
	fmt.Println(strings.Repeat("-", 80))

	periodRecords := engine.ProcessPeriodEnd()

	if len(periodRecords) == 0 {
		fmt.Println("   （暂无周期性奖励数据）")
	} else {
		fmt.Printf("%-6s | %-12s | %10s | %-30s\n",
			"用户ID", "奖励类型", "金额(元)", "备注")
		fmt.Println(strings.Repeat("-", 80))

		periodTotal := 0.0
		for _, r := range periodRecords {
			fmt.Printf("%-6s | %-12s | %10.2f | %-30s\n",
				r.UserID, r.BonusType, r.Amount, r.Remark)
			periodTotal += r.Amount
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("周期奖励总额: %.2f 元\n", periodTotal)
	}
	fmt.Println()

	// ─────────────────────────────────────────────────────────────────────────
	// 6. 总结
	// ─────────────────────────────────────────────────────────────────────────
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("📋 系统说明")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
	fmt.Println("本演示系统包含以下奖励机制：")
	fmt.Println()
	fmt.Println("✅ 实时奖励（订单触发）：")
	fmt.Println("   1. 级差奖     - 基于职级差异的分润")
	fmt.Println("   2. 推荐奖     - 直接推荐人的固定比例奖励")
	fmt.Println("   3. 越级奖     - 第2代及以后的多层奖励")
	fmt.Println("   4. 平级奖     - 同职级间的额外激励")
	fmt.Println("   5. 见点奖     - 每单固定金额奖励")
	fmt.Println()
	fmt.Println("✅ 周期奖励（月结/季结）：")
	fmt.Println("   6. 团队管理奖 - 基于团队总业绩的管理奖励")
	fmt.Println("   7. 感恩奖     - 分享下级奖金的匹配奖励")
	fmt.Println("   8. 全局分红   - 公司总业绩的按权重分配")
	fmt.Println()
	fmt.Println("🔍 其他常见机制（代码中已实现框架）：")
	fmt.Println("   9. 复购奖     - 老客户复购的额外奖励")
	fmt.Println("   10. 育成奖    - 培养下级晋升的奖励")
	fmt.Println("   11. 晋级奖    - 职级提升的一次性奖励")
	fmt.Println("   12. 层奖      - 矩阵制中的按层奖励")
	fmt.Println()
	fmt.Println("📚 更多详情请查看 README.md")
	fmt.Println()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✨ 演示完成！")
	fmt.Println(strings.Repeat("=", 80))
}
