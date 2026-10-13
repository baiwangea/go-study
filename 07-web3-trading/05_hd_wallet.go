//go:build level

// ─────────────────────────────────────────────────────────────────────────────
// 「直观学习模式」：一个文件 = 一关 = 完整 main()，直接 go run 本文件即可。
// 首行 //go:build level 让 go build ./... / go vet ./... 跳过本文件，
// 避免同目录多个 func main() 冲突；显式指定文件名运行时会绕过该约束。
//
// 跑法：cd 07-web3-trading && go run 05_hd_wallet.go
// ─────────────────────────────────────────────────────────────────────────────

package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

/*
═══════════════════════════════════════════════════════════════════════
W-05：BIP-39 助记词与 BIP-44 派生 —— 一个助记词管所有地址
═══════════════════════════════════════════════════════════════════════

【核心概念】

	交易机器人要管的不止一个地址（主钱包、子账户、归集地址……）。
	一个地址一把私钥、抄一张纸，是灾难。行业标准做法是三层：

	    熵 (128/256 bit 随机数)
	        │  BIP-39：查 2048 词表 + 校验和
	        ▼
	    助记词 "abandon abandon ... about"   ← 人类唯一要备份的东西
	        │  BIP-39：PBKDF2-HMAC-SHA512(2048 轮, salt="mnemonic"+口令)
	        ▼
	    512 bit 种子
	        │  BIP-32：HMAC-SHA512 逐层派生
	        ▼
	    主密钥 m
	        │  BIP-44：定义路径语义 m / purpose' / coin_type' / account' / change / index
	        ▼
	    m/44'/60'/0'/0/0  →  0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266

	三个标准各管一段，缺一不可：
	    BIP-39  管「怎么把随机数变成好记好抄的词」——备份格式
	    BIP-32  管「怎么从一把种子无限派生子密钥」——树形结构
	    BIP-44  管「这棵树每个分支叫什么名字」——路径语义

【你会看到】

	✓ 16 字节熵 → 12 个词，32 字节熵 → 24 个词（熵每多 4 字节多 1 个词）
	✓ 同一个助记词跑两遍，派生出的 3 个地址完全一致 —— 这就是「可复现」
	✓ m/44'/60'/0'/0/0、/1、/2 三个地址，和 Hardhat/Anvil 默认账户一模一样
	✓ 同一个助记词，只把 ' 去掉（不硬化），地址就完全不同
	✓ 从「父公钥」能推出「不硬化子公钥」，但推不出「硬化子公钥」

【思考题】（动手改代码验证，别空想）

 1. 把 testMnemonic 的最后两个词互换位置，再跑一遍。
    地址变了，但 bip39.IsMnemonicValid 返回什么？为什么校验和救不了你？
    （提示：校验和只管「这串词是不是一个合法的助记词」，不管「是不是你要的那个」。）

 2. 把 deriveAddress 里的 hardened(44) 改成裸的 44（不硬化），
    地址立刻变化。现在回答：**BIP-44 为什么强制前三层必须硬化？**
    （提示：见示例 5，不硬化的子公钥可以从父公钥推出来。）

 3. 把 NewSeed 的第二个参数从 "" 改成 "TREZOR"（这是 BIP-39 官方测试向量里的
    口令 passphrase）。地址全变了。那么问题来了：
    **备份助记词时忘了备份口令，钱还在不在？**
    （提示：在，但你永远找不回来。这就是「第 25 个词」的坑。）

 4. 用 NewEntropy(16) 和 NewEntropy(32) 各生成一次助记词，数词数。
    为什么是 12/15/18/21/24 而不是随便几个？每个词携带多少 bit？
    （提示：2048 = 2^11，每个词 11 bit。）

 5. 本关的派生路径是 m/44'/60'/0'/0/i。如果机器人要给每个策略分配独立子账户，
    该改路径的哪一段？改了之后「同一策略的所有地址」还能被一个 xpub 管住吗？

【通关标准】

	能手写「助记词 → 种子 → BIP-44 路径 → 地址」的完整链条，
	能说清 44'/60'/0'/0/i 每一段的含义，
	能解释 hardened 与 non-hardened 的安全差别，
	并且知道助记词等于全部资产、绝不能进代码库。

【前置】L1-19（哈希与编码）   【标签】BIP-39 · BIP-32 · BIP-44 · 助记词 · 热钱包
═══════════════════════════════════════════════════════════════════════
*/

// testMnemonic 是公开的测试助记词（Hardhat / Foundry Anvil 的默认账户来源）。
//
// ⚠️ 安全提醒：它之所以能写在源码里，是因为**全世界都知道它**，
// 任何转进去的钱都会被机器人秒扫走。它的唯一用途就是本地测试。
// 真实助记词绝不能出现在代码、日志、git 历史、聊天记录里。
const testMnemonic = "test test test test test test test test test test test junk"

// hardened 把子索引标记为「硬化派生」：把最高位置 1。
//
// BIP-32 规定子索引是 uint32，但只用了低 31 位（0 ~ 2^31-1）表示普通索引；
// 第 32 位（0x80000000 = 2147483648）是「硬化标志位」。
// 所以 44' 在代码里就是 44 | 0x80000000，写作十进制是 2147483692。
//
// 因为标志位和索引位互不重叠，用 | 还是 + 结果一样，
// 但 | 更准确地表达了「设置一个标志位」的语义。
func hardened(index uint32) uint32 {
	return index | bip32.FirstHardenedChild
}

// formatPath 把 []uint32 渲染成人类可读的 "m/44'/60'/0'/0/0"。
//
// 带 ' 的是硬化层。之所以要写这个函数，是因为代码里没有「硬化」这个类型，
// 只有被置了最高位的数字 —— 不渲染出来，你自己都看不出 2147483692 是 44'。
func formatPath(path []uint32) string {
	var b strings.Builder
	b.WriteString("m")
	for _, idx := range path {
		if idx&bip32.FirstHardenedChild != 0 {
			fmt.Fprintf(&b, "/%d'", idx&^bip32.FirstHardenedChild)
		} else {
			fmt.Fprintf(&b, "/%d", idx)
		}
	}
	return b.String()
}

// deriveAddress 沿 BIP-32 路径逐层派生，返回终点私钥对应的以太坊地址。
//
// 三行核心逻辑，正好对应三个标准：
//  1. NewMasterKey(seed)          ← BIP-32 的第一层（HMAC-SHA512("Bitcoin seed", seed)）
//  2. NewChildKey(idx) 循环        ← BIP-32 的逐层派生（每层一次 HMAC-SHA512）
//  3. ToECDSA + PubkeyToAddress   ← secp256k1 公钥 → Keccak256 → 取后 20 字节
//
// 注意第 3 步：以太坊地址**不是**公钥的哈希前缀，而是
// Keccak256(未压缩公钥去掉 0x04 前缀的 64 字节)[后 20 字节]。
// 这一点和比特币不同（比特币是 RIPEMD160(SHA256(pubkey))）。
func deriveAddress(seed []byte, path []uint32) (common.Address, error) {
	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return common.Address{}, fmt.Errorf("生成主密钥失败: %w", err)
	}

	key := master
	for depth, idx := range path {
		key, err = key.NewChildKey(idx)
		if err != nil {
			return common.Address{}, fmt.Errorf("派生第 %d 层（索引 %d）失败: %w", depth+1, idx, err)
		}
	}

	// bip32 的 Key.Key 是 32 字节私钥原始字节，转成 go-ethereum 的 ECDSA 私钥。
	priv, err := crypto.ToECDSA(key.Key)
	if err != nil {
		return common.Address{}, fmt.Errorf("转 ECDSA 私钥失败: %w", err)
	}
	return crypto.PubkeyToAddress(priv.PublicKey), nil
}

// derivePrivateKey 和 deriveAddress 走同一条路径，但额外返回终点私钥，
// 供 W-08 演示「用派生出来的私钥签名」时复用。
func derivePrivateKey(seed []byte, path []uint32) (*bip32.Key, error) {
	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	key := master
	for _, idx := range path {
		key, err = key.NewChildKey(idx)
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}

// wordCount 数助记词里有几个词。BIP-39 的合法词数只有 12/15/18/21/24。
func wordCount(mnemonic string) int {
	return len(strings.Fields(mnemonic))
}

// mustPublic 把 32 字节私钥原始字节转成 ECDSA 公钥，只用于本关的演示。
// 出错就直接 panic —— 测试代码里没必要为「不可能发生」的情况写分支。
func mustPublic(privBytes []byte) ecdsa.PublicKey {
	priv, err := crypto.ToECDSA(privBytes)
	if err != nil {
		panic(fmt.Sprintf("私钥转 ECDSA 失败: %v", err))
	}
	return priv.PublicKey
}

// pubToAddress 把 bip32 的公钥节点转成以太坊地址。
//
// bip32 的公钥是 33 字节**压缩**格式（0x02/0x03 + 32 字节 X 坐标），
// 而 Keccak256 要的是 64 字节未压缩坐标，所以必须先 DecompressPubkey。
// 这个细节是「从 xpub 生成收款地址」时最容易踩的坑。
func pubToAddress(k *bip32.Key) common.Address {
	pub, err := crypto.DecompressPubkey(k.Key)
	if err != nil {
		panic(fmt.Sprintf("解压公钥失败: %v", err))
	}
	return crypto.PubkeyToAddress(*pub)
}

// ethPath 是 BIP-44 定义的以太坊标准路径前缀：m/44'/60'/account'/change
//
//	44' = purpose      ，BIP-44 这个标准本身
//	60' = coin_type    ，SLIP-44 给以太坊分配的是 60（比特币是 0）
//	0'  = account      ，账户序号，用来把资金隔离成互不可见的几组
//	0   = change       ，0 = 外部链（收钱地址），1 = 内部链（找零地址）
//
// 最后再接一个不硬化的 index，就是「第几个地址」。
func ethPath(account, change uint32) []uint32 {
	return []uint32{hardened(44), hardened(60), hardened(account), change}
}

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 W-05 BIP-39 助记词与 BIP-44 派生")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// ── 示例 1：熵 → 助记词（BIP-39 第一步） ─────────────────────────────
	fmt.Println("📝 示例 1：熵 → 助记词")
	fmt.Println(strings.Repeat("-", 60))
	// 先看看真实的随机熵长什么样（长度是随机的，所以只打印长度，不打印内容）。
	// 注意：NewEntropy 的参数单位是 **bit**，不是字节。
	ent, err := bip39.NewEntropy(128) // 128 bit = 16 字节
	if err != nil {
		fmt.Println("  生成熵失败：", err)
		return
	}
	fmt.Printf("  bip39.NewEntropy(128) → %d 字节（%d bit）随机数\n", len(ent), len(ent)*8)

	// 为了输出可复现，下面改用**固定**熵，而不是随机熵。
	// 这是 BIP-39 官方测试向量里的经典例子：全 0 熵。
	zeroEntropy := make([]byte, 16)
	zeroMnemonic, err := bip39.NewMnemonic(zeroEntropy)
	if err != nil {
		fmt.Println("  熵转助记词失败：", err)
		return
	}
	fmt.Printf("  全 0 熵(16B)          → %s\n", zeroMnemonic)
	fmt.Printf("                          词数 = %d\n", wordCount(zeroMnemonic))

	fullEntropy := make([]byte, 32)
	for i := range fullEntropy {
		fullEntropy[i] = 0xff
	}
	fullMnemonic, _ := bip39.NewMnemonic(fullEntropy)
	fmt.Printf("  全 ff 熵(32B)         → %s\n", fullMnemonic)
	fmt.Printf("                          词数 = %d\n", wordCount(fullMnemonic))

	// 为什么 16 字节 → 12 个词？因为 BIP-39 把熵切成 11 bit 一块去查词表：
	//   128 bit 熵 + 4 bit 校验和 = 132 bit = 12 × 11 bit
	//   256 bit 熵 + 8 bit 校验和 = 264 bit = 24 × 11 bit
	// 校验和长度 = 熵长度 / 32，作用是让你抄错一个词时能被立刻发现。
	fmt.Println("  规律：熵每多 4 字节（32 bit），助记词就多 3 个词；校验和 = 熵长度/32 bit")

	// ── 示例 2：助记词 → 种子（BIP-39 第二步） ───────────────────────────
	fmt.Println("📌 示例 2：助记词 → 512 bit 种子")
	fmt.Println(strings.Repeat("-", 60))
	seed := bip39.NewSeed(testMnemonic, "") // 第二个参数是 passphrase，默认空串
	fmt.Printf("  助记词        → %s\n", testMnemonic)
	fmt.Printf("  词数          → %d\n", wordCount(testMnemonic))
	fmt.Printf("  合法性校验    → %t\n", bip39.IsMnemonicValid(testMnemonic))
	// 把最后一个词改掉，校验和立刻不过 —— 这就是校验和的价值。
	fmt.Printf("  末词改成 junk2 → %t（校验和不匹配）\n",
		bip39.IsMnemonicValid("test test test test test test test test test test test junk2"))
	fmt.Printf("  种子          → %s\n", hex.EncodeToString(seed))
	fmt.Printf("  种子长度      → %d 字节（固定 64，和助记词长度无关）\n", len(seed))

	// NewSeed 内部是 PBKDF2-HMAC-SHA512，迭代 2048 轮，salt = "mnemonic" + passphrase。
	// 迭代 2048 轮是故意让暴力破解变慢的 —— 助记词只有 128 bit 熵，
	// 但攻击者还得为每个猜测重跑 2048 轮哈希。
	fmt.Println("  原理：PBKDF2-HMAC-SHA512，2048 轮，salt = \"mnemonic\" + passphrase")
	fmt.Println("  ⚠️ 第二个参数（passphrase）就是俗称的「第 25 个词」：")
	fmt.Println("     备份了助记词却忘了它，钱在链上但永远取不回来。")

	// ── 示例 3：种子 → 地址（BIP-44，可复现） ────────────────────────────
	fmt.Println("🔍 示例 3：按 m/44'/60'/0'/0/i 派生前 3 个地址")
	fmt.Println(strings.Repeat("-", 60))
	base := ethPath(0, 0) // m/44'/60'/0'/0
	fmt.Printf("  路径前缀：%s\n", formatPath(base))
	fmt.Println("  （下面 3 个地址就是 Hardhat / Foundry Anvil 的默认账户 0/1/2）")

	type derived struct {
		index uint32
		addr  common.Address
		path  []uint32
	}
	var got []derived
	for i := uint32(0); i < 3; i++ {
		p := append(append([]uint32{}, base...), i)
		addr, err := deriveAddress(seed, p)
		if err != nil {
			fmt.Println("  派生失败：", err)
			return
		}
		got = append(got, derived{index: i, addr: addr, path: p})
		fmt.Printf("  [%d] %-22s → %s\n", i, formatPath(p), addr.Hex())
	}

	// 可复现性验证：完全重新派生一遍，结果必须逐字节相同。
	// 这是钱包最核心的性质 —— 只要助记词一样，任何机器、任何时候都得到同一个地址。
	fmt.Println("\n  可复现性验证（重新派生一遍再比对）：")
	allSame := true
	for _, d := range got {
		again, _ := deriveAddress(seed, d.path)
		same := again == d.addr
		allSame = allSame && same
		fmt.Printf("    [%d] 二次派生一致 = %t\n", d.index, same)
	}
	fmt.Printf("  → 全部一致 = %t。助记词是「唯一的备份物」，道理就在这里。\n", allSame)

	// 顺带看一眼终点私钥：地址是公开的，私钥才是钱。
	key0, _ := derivePrivateKey(seed, got[0].path)
	fmt.Printf("\n  [0] 的私钥（⚠️ 测试用，公开的）：%s\n", hex.EncodeToString(key0.Key))

	// ── 示例 4：路径每一段的含义 ─────────────────────────────────────────
	fmt.Println("🧪 示例 4：m/44'/60'/0'/0/i 每一段在说什么")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  44'  purpose   → 我遵循 BIP-44 规范（不是 BIP-49/BIP-84）")
	fmt.Println("  60'  coin_type → SLIP-44 给以太坊分配的编号（比特币是 0）")
	fmt.Println("  0'   account   → 第几个账户，用来把资金隔离成互不可见的几组")
	fmt.Println("  0    change    → 0=外部链（对外收款的地址），1=内部链（找零）")
	fmt.Println("  i    index     → 这个链上的第几个地址，从 0 开始递增")
	fmt.Println("  带 ' 的（前三层）是硬化派生，不带的是普通派生。")

	// 换任何一段，地址都不同。下面用实测值证明「每一段都是有效变量」：
	btcPath := []uint32{hardened(44), hardened(0), hardened(0), 0, 0}   // coin_type 换成 BTC 的 0
	acc1Path := []uint32{hardened(44), hardened(60), hardened(1), 0, 0} // account 换成 1
	btcAddr, _ := deriveAddress(seed, btcPath)
	acc1Addr, _ := deriveAddress(seed, acc1Path)
	fmt.Printf("\n  改 coin_type → %-24s = %s（以太坊地址，但这是 BTC 的派生分支）\n",
		formatPath(btcPath), btcAddr.Hex())
	fmt.Printf("  改 account   → %-24s = %s\n", formatPath(acc1Path), acc1Addr.Hex())
	fmt.Println("  同一个助记词能同时管 BTC / ETH / 各账户，靠的就是路径分段。")

	// ── 示例 5：hardened 与 non-hardened 的区别 ──────────────────────────
	fmt.Println("⚙️ 示例 5：hardened（带 '）vs non-hardened")
	fmt.Println(strings.Repeat("-", 60))
	// 先证明「加不加 ' 结果完全不同」。
	softPath := []uint32{44, 60, 0, 0, 0} // 全部不硬化
	softAddr, _ := deriveAddress(seed, softPath)
	fmt.Printf("  %-22s → %s（全不硬化）\n", formatPath(softPath), softAddr.Hex())
	fmt.Printf("  %-22s → %s（全硬化）\n", formatPath([]uint32{hardened(44), hardened(60), hardened(0), 0, 0}), got[0].addr.Hex())

	// 再看安全上的差别：不硬化的子公钥可以由父公钥算出来。
	// 逐层走到 m/44'/60'/0'/0（也就是「外部链」这一层）。
	chainKey, err := derivePrivateKey(seed, base)
	if err != nil {
		fmt.Println("  派生失败：", err)
		return
	}
	fmt.Printf("\n  外部链层的压缩公钥（33 字节）：%x\n", chainKey.PublicKey().Key)

	// ① 从父「私钥」推 0 号子公钥
	fromPriv, _ := chainKey.NewChildKey(0)
	addrFromPriv := crypto.PubkeyToAddress(mustPublic(fromPriv.Key))

	// ② 从父「公钥」直接推 0 号子公钥 —— 没有私钥也能做到！
	fromPub, err := chainKey.PublicKey().NewChildKey(0)
	if err != nil {
		fmt.Println("  公钥派生子键失败：", err)
		return
	}
	addrFromPub := pubToAddress(fromPub)

	fmt.Printf("  从父私钥推 0 号地址：%s\n", addrFromPriv.Hex())
	fmt.Printf("  从父公钥推 0 号地址：%s\n", addrFromPub.Hex())
	fmt.Printf("  两者一致 = %t  ← 这就是「不硬化」的含义\n", addrFromPriv == addrFromPub)

	// ③ 硬化层不能用公钥推。这是硬化的全部意义所在。
	_, errHardened := chainKey.PublicKey().NewChildKey(hardened(1))
	fmt.Printf("  从父公钥推 1'（硬化）子键：%v\n", errHardened)

	fmt.Println("\n  为什么硬化更安全？")
	fmt.Println("    不硬化：只要泄露「父公钥 + 任意一个子私钥」，攻击者就能反推出父私钥，")
	fmt.Println("            进而拿走这棵子树下的全部资产。（所以它只配用来做只读观察钱包）")
	fmt.Println("    硬化  ：派生只用父私钥，公钥这条路根本走不通，")
	fmt.Println("            泄露 xpub 不会泄露任何私钥，也推不出硬化兄弟。")
	fmt.Println("    所以 BIP-44 强制 44'/60'/0' 三层必须硬化，")
	fmt.Println("    只把最末的 change/index 留成不硬化，方便服务端拿 xpub 批量生成收款地址。")

	// ── 示例 6：安全提醒（本关最重要的一段） ─────────────────────────────
	fmt.Println("🎯 示例 6：安全提醒")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("  ✅ 助记词 = 私钥 = 全部资产。拿到助记词的人不需要任何额外信息。")
	fmt.Println("  ✅ 本关的助记词是全公开的测试助记词，绝不能用它存真钱。")
	fmt.Println("  ✅ 助记词绝不能进代码、git 历史、日志、截图、聊天记录、云笔记。")
	fmt.Println("  ✅ 正确做法：环境变量 / KMS / 硬件钱包 / 加密后的 keystore 文件。")
	fmt.Println("  ✅ 交易机器人里的钱叫「热钱包」，只放你亏得起的量。")
	fmt.Println("    大额资产放冷钱包（离线签名），机器人只做小额高频。")
	fmt.Println("  ✅ 派生路径也要一起备份：助记词 + 路径 + passphrase 三样齐了才能找回钱。")
	fmt.Println("  ✅ 对照其他语言：")
	fmt.Println("      PHP  web3.php 用 `Mnemonic` / `HDWallet`，底层同样是 bitcore-lib-php")
	fmt.Println("      Java web3j 用 `Bip39Wallet` + `Credentials.createBip39Wallet(...)`")
	fmt.Println("      Go   go-ethereum 只提供 crypto/keccak/secp256k1，")
	fmt.Println("           BIP-39/32 要靠 tyler-smith/go-bip39 与 go-bip32 补上")

	// ── 关键总结 ─────────────────────────────────────────────────────────
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  熵 --BIP-39--> 助记词 --BIP-39(PBKDF2,2048轮)--> 种子 --BIP-32--> 主密钥")
	fmt.Println("  主密钥 --BIP-44 路径--> m/44'/60'/0'/0/i --> 地址")
	fmt.Println("  ✅ 助记词是唯一备份物：同样助记词 → 永远同样的地址")
	fmt.Println("  ✅ 44'=BIP44 标准，60'=以太坊，0'=账户，0=外部链，i=地址索引")
	fmt.Println("  ✅ 硬化层（带 '）只能由私钥派生；不硬化层公钥也能派生，安全性更低")
	fmt.Println("  ✅ 助记词等于私钥等于资产，热钱包只放小额")
}
