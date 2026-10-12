// PnL 的表驱动测试：毛利、手续费、净利、胜率、最大回撤、期末持仓逐项钉住。
//
// 断言用字符串而不是 Equal()：StringFixed 能同时验证「数值对」和「展示精度对」，
// 如果哪天有人把 decimal 换回 float64，长小数会立刻露馅。

package backtest

import (
	"testing"

	"github.com/shopspring/decimal"
)

// fill 造一笔成交，数量与价格都用字符串给，避免测试自己引入浮点误差。
func fill(side, qty, price string) Fill {
	return Fill{
		OrderID: "T",
		Symbol:  "BTCUSDT",
		Side:    side,
		Qty:     decimal.RequireFromString(qty),
		Price:   decimal.RequireFromString(price),
	}
}

func TestCompute(t *testing.T) {
	tests := []struct {
		name  string
		fills []Fill
		feeBP string

		wantGross   string
		wantFee     string
		wantNet     string
		wantTrades  int
		wantWins    int
		wantWinRate string
		wantDD      string
		wantOpen    string
	}{
		{
			name:  "没有任何成交",
			fills: nil, feeBP: "5",
			wantGross: "0.00", wantFee: "0.00", wantNet: "0.00",
			wantTrades: 0, wantWins: 0, wantWinRate: "0.0", wantDD: "0.00", wantOpen: "0",
		},
		{
			name:  "只开仓不平仓：毛利为 0，只亏手续费",
			fills: []Fill{fill(sideBuy, "1", "100")}, feeBP: "5",
			wantGross: "0.00", wantFee: "0.05", wantNet: "-0.05",
			wantTrades: 0, wantWins: 0, wantWinRate: "0.0", wantDD: "0.05", wantOpen: "1",
		},
		{
			name:  "一开一平盈利（免手续费）",
			fills: []Fill{fill(sideBuy, "1", "100"), fill(sideSell, "1", "110")}, feeBP: "0",
			wantGross: "10.00", wantFee: "0.00", wantNet: "10.00",
			wantTrades: 1, wantWins: 1, wantWinRate: "100.0", wantDD: "0.00", wantOpen: "0",
		},
		{
			name:  "一开一平亏损（免手续费）：回撤等于亏损额",
			fills: []Fill{fill(sideBuy, "1", "100"), fill(sideSell, "1", "90")}, feeBP: "0",
			wantGross: "-10.00", wantFee: "0.00", wantNet: "-10.00",
			wantTrades: 1, wantWins: 0, wantWinRate: "0.0", wantDD: "10.00", wantOpen: "0",
		},
		{
			name:  "做空盈利：持仓为负时价格跌才算赚",
			fills: []Fill{fill(sideSell, "1", "100"), fill(sideBuy, "1", "90")}, feeBP: "0",
			wantGross: "10.00", wantFee: "0.00", wantNet: "10.00",
			wantTrades: 1, wantWins: 1, wantWinRate: "100.0", wantDD: "0.00", wantOpen: "0",
		},
		{
			name:  "手续费按每笔成交的名义金额双边收取",
			fills: []Fill{fill(sideBuy, "1", "100"), fill(sideSell, "1", "110")}, feeBP: "10",
			// 0.1（买入 100×10bp）+ 0.11（卖出 110×10bp）= 0.21
			wantGross: "10.00", wantFee: "0.21", wantNet: "9.79",
			wantTrades: 1, wantWins: 1, wantWinRate: "100.0", wantDD: "0.10", wantOpen: "0",
		},
		{
			name: "同向加仓按成交量加权开仓均价",
			fills: []Fill{
				fill(sideBuy, "1", "100"),
				fill(sideBuy, "1", "120"), // 均价 (100+120)/2 = 110
				fill(sideSell, "2", "130"),
			}, feeBP: "0",
			wantGross: "40.00", wantFee: "0.00", wantNet: "40.00",
			wantTrades: 1, wantWins: 1, wantWinRate: "100.0", wantDD: "0.00", wantOpen: "0",
		},
		{
			name:  "部分平仓：只对平掉的部分实现盈亏，剩余仓位留着",
			fills: []Fill{fill(sideBuy, "2", "100"), fill(sideSell, "1", "120")}, feeBP: "0",
			wantGross: "20.00", wantFee: "0.00", wantNet: "20.00",
			wantTrades: 1, wantWins: 1, wantWinRate: "100.0", wantDD: "0.00", wantOpen: "1",
		},
		{
			name: "反手：先平完再反向开新仓，新仓开仓价是本次成交价",
			fills: []Fill{
				fill(sideBuy, "1", "100"),
				fill(sideSell, "3", "110"), // 平掉 1 手赚 10，剩 −2 手空单，开仓价 110
				fill(sideBuy, "2", "100"),  // 平掉 2 手空单：(100−110)×2×(−1) = +20
			}, feeBP: "0",
			wantGross: "30.00", wantFee: "0.00", wantNet: "30.00",
			wantTrades: 2, wantWins: 2, wantWinRate: "100.0", wantDD: "0.00", wantOpen: "0",
		},
		{
			name: "胜率 = 盈利平仓笔数 / 总平仓笔数",
			fills: []Fill{
				fill(sideBuy, "1", "100"), fill(sideSell, "1", "110"), // 赚
				fill(sideSell, "1", "110"), fill(sideBuy, "1", "100"), // 赚
				fill(sideBuy, "1", "100"), fill(sideSell, "1", "90"), // 亏
			}, feeBP: "0",
			wantGross: "10.00", wantFee: "0.00", wantNet: "10.00",
			wantTrades: 3, wantWins: 2, wantWinRate: "66.7", wantDD: "10.00", wantOpen: "0",
		},
		{
			name: "最大回撤取已实现净权益曲线的峰谷差",
			fills: []Fill{
				fill(sideBuy, "1", "100"), fill(sideSell, "1", "90"), // 权益 0 → −10，回撤 10
				fill(sideBuy, "1", "90"), fill(sideSell, "1", "110"), // 权益 → +10，回撤归 0
			}, feeBP: "0",
			wantGross: "10.00", wantFee: "0.00", wantNet: "10.00",
			wantTrades: 2, wantWins: 1, wantWinRate: "50.0", wantDD: "10.00", wantOpen: "0",
		},
		{
			name: "手续费会加深回撤",
			fills: []Fill{
				fill(sideBuy, "1", "100"), fill(sideSell, "1", "90"),
			}, feeBP: "10",
			// 亏 10 + 手续费 0.1 + 0.09 = 0.19
			wantGross: "-10.00", wantFee: "0.19", wantNet: "-10.19",
			wantTrades: 1, wantWins: 0, wantWinRate: "0.0", wantDD: "10.19", wantOpen: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compute(tt.fills, decimal.RequireFromString(tt.feeBP))

			checks := []struct {
				name string
				got  string
				want string
			}{
				{"毛利", got.Gross.StringFixed(2), tt.wantGross},
				{"手续费", got.Fee.StringFixed(2), tt.wantFee},
				{"净利", got.Net.StringFixed(2), tt.wantNet},
				{"胜率", got.WinRate.StringFixed(1), tt.wantWinRate},
				{"最大回撤", got.MaxDrawdown.StringFixed(2), tt.wantDD},
				{"期末持仓", got.OpenQty.String(), tt.wantOpen},
			}
			for _, c := range checks {
				if c.got != c.want {
					t.Errorf("%s = %s, want %s", c.name, c.got, c.want)
				}
			}
			if got.Fills != len(tt.fills) {
				t.Errorf("成交笔数 = %d, want %d", got.Fills, len(tt.fills))
			}
			if got.ClosedTrades != tt.wantTrades {
				t.Errorf("平仓笔数 = %d, want %d", got.ClosedTrades, tt.wantTrades)
			}
			if got.Wins != tt.wantWins {
				t.Errorf("盈利笔数 = %d, want %d", got.Wins, tt.wantWins)
			}
		})
	}
}

// 金额全程 decimal：0.3 / 0.1 / 0.2 这类十进制小数在 float64 里是无限循环小数，
// 一旦中间走过 float64，末位就会出现 ...9999996 这种尾巴。这里直接比字符串。
func TestComputeExactDecimal(t *testing.T) {
	fills := []Fill{
		fill(sideBuy, "0.3", "63000.1"),
		fill(sideSell, "0.1", "63111.7"),
		fill(sideSell, "0.2", "62999.3"),
	}
	s := Compute(fills, decimal.RequireFromString("7.5"))

	// 毛利：(63111.7−63000.1)×0.1 + (62999.3−63000.1)×0.2 = 11.16 − 0.16 = 11
	// 手续费：0.3×63000.1×7.5/1e4 + 0.1×63111.7×7.5/1e4 + 0.2×62999.3×7.5/1e4
	//       = 14.1750225 + 4.7333775 + 9.449895 = 28.358295
	tests := []struct {
		name string
		got  decimal.Decimal
		want string
	}{
		{"毛利", s.Gross, "11"},
		{"手续费", s.Fee, "28.358295"},
		{"净利", s.Net, "-17.358295"},
	}
	for _, tt := range tests {
		if tt.got.String() != tt.want {
			t.Errorf("%s = %s, want %s（末位有尾巴说明中间走了 float64）", tt.name, tt.got, tt.want)
		}
	}
	if want := s.Gross.Sub(s.Fee); !s.Net.Equal(want) {
		t.Errorf("净利 = %s, want 毛利−手续费 = %s", s.Net, want)
	}
}
