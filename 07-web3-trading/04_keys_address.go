//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 04_keys_address.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

/*
═══════════════════════════════════════════════════════════════════════
W-04：私钥 → 地址 与 keystore
═══════════════════════════════════════════════════════════════════════

【核心概念】
    前四关讲的是「钱怎么表示」，这一关开始讲「链上身份」。

    一条单向推导链，走完就回不去了：
        私钥(32 字节随机数) → 公钥(secp256k1 椭圆曲线乘法) → 地址(Keccak256 + 截断)
    每一步都不可逆：从地址推不出公钥，从公钥推不出私钥。
    所以「地址可以公开，私钥泄露 = 资产全丢」。

    具体算法（**背下来，面试和排错都用得上**）：
        ① 私钥是 32 字节（256 位）随机数，必须落在 secp256k1 的合法区间内
        ② 公钥 = 曲线上的点 (X, Y)，序列化成 65 字节：1 字节 0x04 前缀 + 32 字节 X + 32 字节 Y
        ③ 把前缀去掉，对**剩下的 64 字节**做 Keccak256 → 32 字节哈希
        ④ 取哈希的**后 20 字节** → 这就是地址

    ⚠️ 三个最常见的错误，本关都会让你亲眼看到错出来的地址长什么样：
        · 用了 SHA-256（那是比特币的，以太坊用的是 Keccak-256，两者**不是**同一个算法）
        · 取了哈希的**前** 20 字节（应该取后 20）
        · 把带 0x04 前缀的 65 字节一起哈希（应该去掉前缀）

    地址的十六进制字符串还带一层校验：EIP-55 用**大小写**编码 4 bit 的校验和。
    所以 0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe 里的大小写不是随便写的，
    它是这个地址的固有属性 —— 打错一个字母的大小写，就说明你抄错了地址。

【你会看到】
    ✓ 私钥 32 字节、公钥 65 字节（0x04 + 64）、地址 20 字节的完整推导过程
    ✓ 用 SHA-256 算出来的「假地址」和取前 20 字节算出来的「假地址」都跟真地址不同
    ✓ 已知向量验证：私钥 0x…0001 → 地址 0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf
    ✓ 手写一遍 EIP-55 算法，结果和 common.Address.Hex() 完全一致
    ✓ keystore 把私钥用密码加密成 JSON：NewAccount → Export → Import → Unlock → 签名
    ✓ 用错密码导出会报 "could not decrypt key with given password"

【思考题】（动手改代码验证，别空想）
    1. 示例 3 里我故意用 SHA-256 和「取前 20 字节」各算了一个假地址。
       请把它们和真地址并排贴出来 —— 如果这类 bug 上线了，用户的币会打到哪去？
       （提示：地址不存在的概率接近 100%，转账会直接失败或永久锁死。）
    2. 示例 4 的手写 EIP-55 和 common.Address.Hex() 输出一致。请解释：
       为什么要用大小写这种「看起来像 bug」的方式做校验和？
       （提示：地址本身是二进制，人类抄写时最怕的是「多一个字符/少一个字符」还是
        「某个字符打错」？大小写在复制粘贴时会不会丢失？）
    3. **本关最重要的一题**：示例 4 末尾演示了 common.HexToAddress 会**静默接受**
       一个校验和错误的地址。请设计一个校验函数，在接收用户输入的提现地址时
       拦住这种错误，并说明为什么这比「检查长度是不是 42」重要得多。
    4. 示例 5 里 NewAccount 用的是 LightScryptN（实测 ~83ms），
       换成 StandardScryptN 后实测 ~1.06s（本机 Apple Silicon）。
       为什么标准模式要故意做这么慢？「慢」在这里保护的是什么？
       （提示：攻击者拿到 keystore JSON 后要做什么？）
    5. 跨语言对照：PHP 的 web3.php、Java 的 web3j、Python 的 eth-account
       都能做同样的推导。请思考：为什么这类**签名/派生**逻辑一定要用成熟的库，
       而不是自己照着白皮书实现一遍？（提示：侧信道、随机数、边界条件。）
    6. 示例 6 的 maskSecret 把私钥中间遮住了。请列出你的机器人在哪些地方
       可能不小心把私钥写进日志/错误信息/panic 栈/HTTP 请求体，
       并各给一条防御措施。

【通关标准】
    能手写出「私钥 → 公钥 → Keccak256 → 后 20 字节 → 地址」的完整链路，
    能解释 EIP-55 校验和为什么用大小写，
    能说出 keystore 是什么、为什么它比明文私钥安全、以及它**并不**等于安全保管方案。

【前置】W-01、W-02（decimal）      【标签】secp256k1 · Keccak256 · EIP-55 · keystore · 私钥安全
═══════════════════════════════════════════════════════════════════════
*/

// deriveAddress 走完「私钥 → 地址」的完整链路，并把中间结果打印出来。
//
// 这是本关的核心，每一步都对应上面【核心概念】里的一条：
//
//	priv.D              → 私钥那个大整数
//	FromECDSAPub        → 65 字节（0x04 前缀 + X + Y）
//	pub[1:]             → 64 字节（去掉前缀）
//	Keccak256(pub[1:])  → 32 字节哈希
//	hash[12:]           → 后 20 字节 = 地址
func deriveAddress(priv *ecdsa.PrivateKey) common.Address {
	// ① 私钥本体：crypto.FromECDSA 把它序列化成 32 字节大端
	privBytes := crypto.FromECDSA(priv)
	fmt.Printf("    ① 私钥           %d 字节  %s\n", len(privBytes), maskSecret(hex.EncodeToString(privBytes)))
	fmt.Printf("       priv.D.BitLen() = %d（合法私钥应接近 256 位）\n", priv.D.BitLen())

	// ② 公钥：未压缩格式，65 字节，第一个字节固定是 0x04
	pub := crypto.FromECDSAPub(&priv.PublicKey)
	fmt.Printf("    ② 公钥(未压缩)   %d 字节  首字节 0x%02x（0x04 = 未压缩标记）\n", len(pub), pub[0])

	// ③ 去掉 0x04 前缀，只对 64 字节做 Keccak256
	//    ⚠️ 这里是最容易错的一步：哈希的输入**不包含**那个 0x04
	pubBody := pub[1:]
	fmt.Printf("    ③ 去前缀后        %d 字节\n", len(pubBody))

	// ④ Keccak256 是「以太坊的 SHA3」，和标准库的 SHA-256 / SHA3-256 都不是一回事
	hash := crypto.Keccak256(pubBody)
	fmt.Printf("    ④ Keccak256      %d 字节  %s\n", len(hash), hex.EncodeToString(hash))

	// ⑤ 取**后** 20 字节。前 12 字节直接丢掉
	addr := common.BytesToAddress(hash[12:])
	fmt.Printf("    ⑤ 取后 20 字节    %d 字节  %s\n", len(addr), addr.Hex())
	return addr
}

// eip55 手写一遍 EIP-55 校验和算法，用来和 common.Address.Hex() 对照。
//
// 算法（非常短，值得背）：
//  1. 取地址的小写 hex 字符串（不带 0x），对它做 Keccak256 → 32 字节哈希
//  2. 逐字符遍历小写 hex：第 i 个字符对应哈希的第 i/2 字节的第 (i%2 ? 低 4 位 : 高 4 位)
//  3. 如果该字符是 a-f，且对应的 4 bit >= 8，就把它转成大写
//  4. 其余字符（0-9）不变
//
// 效果：同一个地址**永远**得到同一个大小写模式；抄错一个字符，
// 大小写模式就对不上，钱包就能立刻报「地址校验和错误」。
func eip55(addr common.Address) string {
	lower := hex.EncodeToString(addr.Bytes()) // 40 个字符，全小写，无 0x
	h := crypto.Keccak256([]byte(lower))
	out := []byte(lower)
	for i, c := range out {
		if c < 'a' || c > 'f' { // 只有 a-f 参与大小写编码
			continue
		}
		var nibble byte
		if i%2 == 0 {
			nibble = h[i/2] >> 4 // 偶数位取哈希字节的高 4 位
		} else {
			nibble = h[i/2] & 0x0f // 奇数位取低 4 位
		}
		if nibble >= 8 {
			out[i] = c - 32 // 转大写（ASCII 里小写比大写大 32）
		}
	}
	return "0x" + string(out)
}

// ValidChecksum 校验一个地址字符串的 EIP-55 校验和是否正确。
//
// 约定（和主流钱包一致）：
//
//	· 全小写 / 全大写 → 视为「未带校验和」，接受（很多工具和旧接口就这么输出）
//	· 混合大小写     → 必须和 EIP-55 算出来的完全一致，否则拒绝
//
// ⚠️ 这就是 common.HexToAddress 的坑：它**只做 hex 解析，不校验大小写**。
//
//	用户把 0xde0B29… 抄成 0xde0b29… 时它能解析成功，
//	但算出来的地址和你以为的可能是两回事 —— 提现地址这种场景必须自己校验。
func ValidChecksum(s string) bool {
	if !common.IsHexAddress(s) {
		return false
	}
	addr := common.HexToAddress(s)
	// 全小写 / 全大写：无校验和信息，只能接受
	if s == "0x"+hex.EncodeToString(addr.Bytes()) || s == "0x"+strings.ToUpper(hex.EncodeToString(addr.Bytes())) {
		return true
	}
	// 混合大小写：必须和 EIP-55 一致
	return s == eip55(addr)
}

// maskSecret 只保留字符串的头 8 位和尾 6 位，中间用 … 遮住。
//
// 这里刻意**不保留原长度**：把「32 字节 / 64 个 hex 字符」这种信息也一起藏掉，
// 免得日志里的星号数量反过来泄露密钥长度。
//
// 这是**唯一**允许私钥出现在日志里的形式。生产代码里建议直接不打印：
// 一旦某次 panic 把整个结构体 %+v 打出来，私钥就进了日志收集系统。
func maskSecret(s string) string {
	if len(s) <= 16 {
		return "******"
	}
	return s[:8] + "…" + s[len(s)-6:]
}

func main() {
	fmt.Println("\n=== W-04 私钥 → 地址 与 keystore ===")

	// ── 示例 1：生成私钥 / 从 hex 恢复 ──────────────────────────────────
	fmt.Println("\n【示例 1】私钥就是一个 32 字节的随机数")
	// crypto.GenerateKey 用 crypto/rand 生成 secp256k1 私钥。
	// 返回值是 *ecdsa.PrivateKey —— Go 标准库的 ecdsa 类型，geth 复用了它。
	priv, err := crypto.GenerateKey()
	if err != nil {
		fmt.Println("  生成失败：", err)
		return
	}
	fmt.Println("  crypto.GenerateKey() 成功，本次的推导过程：")
	generated := deriveAddress(priv)
	fmt.Println("  ⚠️  每次运行这个私钥都不同，所以上面的地址也不同 —— 这是正常的")
	fmt.Println("  ⚠️  真实私钥**绝不能**打印到日志。这里只是演示，用的是没有余额的临时密钥")

	// 从 hex 恢复：HexToECDSA 是「从字符串还原私钥」的唯一入口
	privHex := hex.EncodeToString(crypto.FromECDSA(priv))
	back, err := crypto.HexToECDSA(privHex)
	if err != nil {
		fmt.Println("  HexToECDSA 失败：", err)
		return
	}
	fmt.Printf("  往返验证：HexToECDSA(hex) 的地址 == 原地址 ? %t\n",
		crypto.PubkeyToAddress(back.PublicKey) == generated)
	fmt.Println("  ⚠️  HexToECDSA **不接受** 0x 前缀，传进去会报 invalid hex character 'x'")

	// 用已知向量验证整条链路（这些地址是固定的，可以对照区块浏览器）
	fmt.Println("\n  用已知私钥验证推导链路（结果固定，可自行搜索核对）：")
	for _, k := range []string{
		"0000000000000000000000000000000000000000000000000000000000000001",
		"4646464646464646464646464646464646464646464646464646464646464646",
	} {
		p, err := crypto.HexToECDSA(k)
		if err != nil {
			fmt.Println("  解析失败：", err)
			continue
		}
		fmt.Printf("    私钥 0x%s…%s → %s\n", k[:8], k[len(k)-6:], crypto.PubkeyToAddress(p.PublicKey).Hex())
	}

	// ── 示例 2：公钥是 64 字节还是 65 字节？ ────────────────────────────
	fmt.Println("\n【示例 2】公钥的两种格式：65 字节和 64 字节")
	pubFull := crypto.FromECDSAPub(&priv.PublicKey)
	fmt.Printf("  FromECDSAPub(&priv.PublicKey) = %d 字节（0x04 + 32 字节 X + 32 字节 Y）\n", len(pubFull))
	fmt.Printf("    首字节        = 0x%02x（未压缩标记，恒定）\n", pubFull[0])
	fmt.Printf("    X 坐标        = %s…（%d bit）\n", hex.EncodeToString(pubFull[1:9]), priv.PublicKey.X.BitLen())
	fmt.Printf("    Y 坐标        = %s…（%d bit）\n", hex.EncodeToString(pubFull[33:41]), priv.PublicKey.Y.BitLen())
	fmt.Printf("  去掉首字节后    = %d 字节 ← 这才是做地址哈希的输入\n", len(pubFull[1:]))
	compressed := crypto.CompressPubkey(&priv.PublicKey)
	fmt.Printf("  CompressPubkey = %d 字节（0x02/0x03 + 32 字节 X，只存 X 和 Y 的奇偶性）\n", len(compressed))
	fmt.Println("  ⚠️  算地址要用**未压缩**格式去掉 0x04 后的 64 字节，别拿压缩公钥去算")

	// ── 示例 3：Keccak256 + 取后 20 字节 ────────────────────────────────
	fmt.Println("\n【示例 3】地址 = Keccak256(64 字节公钥) 的后 20 字节")
	body := pubFull[1:]
	correct := crypto.Keccak256(body)
	fmt.Printf("  正确做法    keccak256(pub[1:])  = %s\n", hex.EncodeToString(correct))
	fmt.Printf("              → 取后 20 字节       = %s   ✅\n", common.BytesToAddress(correct[12:]).Hex())

	// 错误做法 A：用了 SHA-256
	wrongA := sha256.Sum256(body)
	fmt.Printf("  错误 A      sha256(pub[1:])     = %s\n", hex.EncodeToString(wrongA[:]))
	fmt.Printf("              → 取后 20 字节       = %s   ❌（比特币用 SHA-256，以太坊用 Keccak-256）\n",
		common.BytesToAddress(wrongA[12:]).Hex())

	// 错误做法 B：取了前 20 字节
	fmt.Printf("  错误 B      取哈希的**前** 20 字节 = %s   ❌（应该取后 20）\n",
		common.BytesToAddress(correct[:20]).Hex())

	// 错误做法 C：把 0x04 前缀也一起哈希了
	wrongC := crypto.Keccak256(pubFull)
	fmt.Printf("  错误 C      keccak256(pub) 整 65 字节 = %s   ❌（多哈希了 0x04）\n",
		common.BytesToAddress(wrongC[12:]).Hex())
	fmt.Println("  四种结果全都不一样。地址算错不会报错，只会把钱打到「别人的」地址上")

	// ── 示例 4：EIP-55 校验和 ───────────────────────────────────────────
	fmt.Println("\n【示例 4】EIP-55：用大小写编码校验和")
	fmt.Printf("  同一个地址的三种写法：\n")
	fmt.Printf("    全小写（无校验和） : 0x%s\n", hex.EncodeToString(generated.Bytes()))
	fmt.Printf("    common.Address.Hex(): %s   ← 混合大小写，大小写**有含义**\n", generated.Hex())
	fmt.Printf("    手写 eip55()        : %s   ← 和上面一致 ? %t\n",
		eip55(generated), eip55(generated) == generated.Hex())

	fmt.Println("\n  用 EIP-55 官方的测试向量验证手写实现（这些地址直接抄自 EIP-55 规范）：")
	for _, s := range []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
		"0x52908400098527886E0F7030069857D2E4169EE7",
		"0x8617E340B3D01FA5F11F306F4090FD50E238070D",
		"0xde709f2102306220921060314715629080e2fb77",
		"0x27b1fdb04752bbc536007a920d24acb045561c26",
	} {
		a := common.HexToAddress(s)
		fmt.Printf("    %s  Hex()==eip55() ? %t   ValidChecksum ? %t\n",
			s, a.Hex() == eip55(a), ValidChecksum(s))
	}

	// 关键演示：HexToAddress 不校验大小写
	fmt.Println("\n  ⚠️  最危险的一点：common.HexToAddress **不校验** EIP-55 大小写")
	good := "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe" // 正确的 EIP-55 形式
	// 把第 4 个字符 B 改成小写 b —— 校验和就错了，但 hex 依然合法
	bad := "0xde0b295669a9FD93d5F28D9Ec85E40f4cb697BAe"
	fmt.Printf("    正确形式 %s\n", good)
	fmt.Printf("    打错大小写 %s\n", bad)
	fmt.Printf("    HexToAddress(打错的) 能否解析 ? %t   ← 能！它只解析 hex，不看大小写\n",
		common.IsHexAddress(bad))
	fmt.Printf("    两者解析出的 Address 相等 ? %t   ← 字节相同，因为地址字节没变\n",
		common.HexToAddress(good) == common.HexToAddress(bad))
	fmt.Printf("    ValidChecksum(正确) ? %t   ValidChecksum(打错) ? %t   ← 只有这个能拦住\n",
		ValidChecksum(good), ValidChecksum(bad))
	fmt.Println("    ↑ 注意这里「打错大小写」并没有改变地址本身，它只是暴露了「你可能抄错了」这个信号")
	fmt.Println("      真正的风险是：用户抄地址时把某个字符抄成了**别的字符**，")
	fmt.Println("      这时 HexToAddress 会给你一个**完全不同的、合法的**地址，钱就转走了")

	// ── 示例 5：keystore 用密码加密私钥 ─────────────────────────────────
	fmt.Println("\n【示例 5】keystore：用密码把私钥加密成 JSON 文件")
	// 用 os.MkdirTemp 建临时目录，函数结束就删掉。
	// 生产代码里这个目录通常是 ~/.ethereum/keystore 或交易所的密钥目录，
	// 而且**必须**是 0700 权限、不进版本控制、不进镜像。
	dir, err := os.MkdirTemp("", "w04-keystore-*")
	if err != nil {
		fmt.Println("  建临时目录失败：", err)
		return
	}
	defer os.RemoveAll(dir)
	fmt.Printf("  临时目录：%s\n", dir)

	// LightScryptN=4096 / StandardScryptN=262144。
	// 本机实测：Light ~83ms，Standard ~1.06s。
	// 生产环境用 Standard —— 慢是**故意的**，它让暴力破解的代价高到不可接受。
	ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

	const passphrase = "correct horse battery staple"
	acct, err := ks.NewAccount(passphrase)
	if err != nil {
		fmt.Println("  NewAccount 失败：", err)
		return
	}
	fmt.Printf("  NewAccount 生成新私钥并加密落盘\n")
	fmt.Printf("    地址 = %s\n", acct.Address.Hex())
	fmt.Printf("    文件 = %s\n", acct.URL.Path)
	fmt.Println("    ⚠️  注意地址和示例 1 那个不一样：NewAccount 会**新生成**一个私钥，")
	fmt.Println("        它跟 priv 没有任何关系。keystore 里存的是它自己生成的那把")

	// 导出的 JSON 里能看到什么
	keyJSON, err := ks.Export(acct, passphrase, passphrase)
	if err != nil {
		fmt.Println("  Export 失败：", err)
		return
	}
	fmt.Printf("  导出 JSON %d 字节，前 160 个字符：\n    %s\n", len(keyJSON), keyJSON[:160])
	fmt.Println("  结构说明（这就是 Web3 Secret Storage 标准）：")
	fmt.Println("    address  = 明文地址（不是秘密）")
	fmt.Println("    crypto   = aes-128-ctr 密文 + scrypt 派生的密钥 + MAC")
	fmt.Println("    → 私钥**没有**明文出现在文件里，但密码强度决定一切：")
	fmt.Println("      弱密码 + StandardScrypt 也只是拖延时间，不是安全")

	// 导入到另一个目录，验证往返
	dir2, err := os.MkdirTemp("", "w04-keystore2-*")
	if err != nil {
		fmt.Println("  建第二个临时目录失败：", err)
		return
	}
	defer os.RemoveAll(dir2)
	ks2 := keystore.NewKeyStore(dir2, keystore.LightScryptN, keystore.LightScryptP)
	imported, err := ks2.Import(keyJSON, passphrase, passphrase)
	if err != nil {
		fmt.Println("  Import 失败：", err)
		return
	}
	fmt.Printf("  导入到另一个目录：地址 %s，和原地址一致 ? %t\n",
		imported.Address.Hex(), imported.Address == acct.Address)

	// 解锁 → 签名 → 用签名恢复公钥，证明「解锁出来的确实是那把私钥」
	if err := ks.Unlock(acct, passphrase); err != nil {
		fmt.Println("  Unlock 失败：", err)
		return
	}
	defer ks.Lock(acct.Address)
	msg := []byte("hello web3")
	msgHash := crypto.Keccak256(msg)
	sig, err := ks.SignHash(acct, msgHash)
	if err != nil {
		fmt.Println("  SignHash 失败：", err)
		return
	}
	fmt.Printf("  签名 %d 字节（以太坊格式 R||S||V，末尾的 V 用于恢复公钥）\n", len(sig))
	recovered, err := crypto.SigToPub(msgHash, sig)
	if err != nil {
		fmt.Println("  SigToPub 失败：", err)
		return
	}
	fmt.Printf("  从签名恢复出的地址 = %s\n", crypto.PubkeyToAddress(*recovered).Hex())
	fmt.Printf("  和原地址一致 ? %t   ← 签名能反推地址、但反推不出私钥，这就是身份认证的基础\n",
		crypto.PubkeyToAddress(*recovered) == acct.Address)

	// 错误密码
	_, err = ks.Export(acct, "wrong password", "x")
	fmt.Printf("  用错密码导出 → %v\n", err)
	fmt.Println("  ↑ 报错是「解密失败」而不是「密码错误」—— 因为密码学上无法区分这两者，")
	fmt.Println("    keystore 只能告诉你 MAC 校验没过")

	// ── 示例 6：私钥安全 ───────────────────────────────────────────────
	fmt.Println("\n【示例 6】私钥安全：这些事一件都不能做")
	fmt.Println("  ❌ 硬编码进代码：")
	fmt.Println("       privHex := \"4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318\"")
	fmt.Println("       → git 历史里永久留存，删了也没用，必须视为已泄露并立刻转移资产")
	fmt.Println("  ❌ 写进配置文件提交仓库：.env / config.yaml / docker-compose.yml 都算")
	fmt.Println("  ❌ 打进日志：log.Printf(\"priv=%s\", privHex) / fmt.Printf(\"%+v\", key)")
	fmt.Println("  ❌ 进 panic 栈：自定义结构体实现了 String()/GoString() 且带了私钥字段")
	fmt.Println("  ❌ 通过 HTTP 传给别的服务（哪怕是自己内网的「签名服务」）")
	fmt.Println("  ✅ 正确做法：")
	fmt.Println("       · keystore 加密文件 + 强密码，密码从环境变量/密钥管理服务注入")
	fmt.Println("       · 或硬件钱包 / HSM / KMS，私钥永不出设备，只把待签名的哈希送进去")
	fmt.Println("       · 机器人用**热钱包**只放运营资金，大额资产放冷钱包")
	fmt.Printf("  ✅ 日志里只允许出现这种形式：%s\n", maskSecret(hex.EncodeToString(crypto.FromECDSA(priv))))
	fmt.Println("     （本关为了演示才把 32 字节私钥打印出来，且这把私钥是随机的、没有余额）")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n【关键总结】")
	fmt.Println("  • 私钥 32 字节随机数 → 公钥 65 字节（0x04+X+Y）→ Keccak256 → 后 20 字节 = 地址")
	fmt.Println("  • 哈希输入是**去掉 0x04 后的 64 字节**；用 SHA-256、取前 20 字节都是错的")
	fmt.Println("  • 链路单向：地址推不出公钥，公钥推不出私钥。私钥泄露 = 资产全丢")
	fmt.Println("  • EIP-55 用大小写编码校验和；common.HexToAddress **不校验**它，要自己比 addr.Hex()")
	fmt.Println("  • keystore = scrypt 派生密钥 + aes-128-ctr 加密 + MAC，导出成 JSON；")
	fmt.Println("    Light ~83ms / Standard ~1.06s（实测），慢是故意的；密码强度决定安全性")
	fmt.Println("  • SignHash 得到 65 字节签名，SigToPub 能从签名恢复地址 —— 认证靠这个，不靠传私钥")
	fmt.Println("  • 私钥绝不硬编码、不提交、不打印、不进 panic 栈；日志里只能出现掩码形式")
}
