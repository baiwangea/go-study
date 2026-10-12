//go:build level

package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	fmt.Println("=== 已知私钥向量 ===")
	for _, k := range []string{
		"0000000000000000000000000000000000000000000000000000000000000001",
		"4646464646464646464646464646464646464646464646464646464646464646",
	} {
		p, err := crypto.HexToECDSA(k)
		if err != nil {
			panic(err)
		}
		addr := crypto.PubkeyToAddress(p.PublicKey)
		fmt.Printf("  私钥 %s...\n    → %s\n", k[:8], addr.Hex())
	}

	fmt.Println("\n=== 0x 前缀 ===")
	_, err := crypto.HexToECDSA("0x0000000000000000000000000000000000000000000000000000000000000001")
	fmt.Println("带 0x 的错误:", err)

	fmt.Println("\n=== Standard vs Light scrypt 耗时 ===")
	for _, cfg := range []struct {
		name string
		n, p int
	}{
		{"Light", keystore.LightScryptN, keystore.LightScryptP},
		{"Standard", keystore.StandardScryptN, keystore.StandardScryptP},
	} {
		dir, _ := os.MkdirTemp("", "ks-*")
		ks := keystore.NewKeyStore(dir, cfg.n, cfg.p)
		st := time.Now()
		_, err := ks.NewAccount("pw")
		fmt.Printf("  %-9s N=%-8d P=%d 耗时 %v\n", cfg.name, cfg.n, cfg.p, time.Since(st))
		if err != nil {
			panic(err)
		}
		os.RemoveAll(dir)
	}

	fmt.Println("\n=== EIP-55 手写实现验证 ===")
	for _, s := range []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
		"0x52908400098527886E0F7030069857D2E4169EE7",
		"0x8617E340B3D01FA5F11F306F4090FD50E238070D",
		"0xde709f2102306220921060314715629080e2fb77",
		"0x27b1fdb04752bbc536007a920d24acb045561c26",
		"0x6F710856Da931Cb8afc1759d75Cb0C6cf950d3aB",
	} {
		a := common.HexToAddress(s)
		mine := manualChecksum(a)
		ok := mine == a.Hex()
		fmt.Printf("  %s  geth=%s  mine=%s  match=%v\n", s, a.Hex(), mine, ok)
	}

	fmt.Println("\n=== 地址字符串相等 ===")
	lower := strings.ToLower("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
	fmt.Println("小写字符串 == 校验和字符串:", lower == "0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
	fmt.Println("HexToAddress(小写).Hex() :", common.HexToAddress(lower).Hex())
	fmt.Println("HexToAddress 比较字节:", common.HexToAddress(lower) == common.HexToAddress("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe"))
}

// manualChecksum 手写 EIP-55
func manualChecksum(a common.Address) string {
	lower := hex.EncodeToString(a.Bytes())
	h := crypto.Keccak256([]byte(lower))
	out := []byte(lower)
	for i, c := range out {
		if c < 'a' || c > 'f' {
			continue
		}
		var nibble byte
		if i%2 == 0 {
			nibble = h[i/2] >> 4
		} else {
			nibble = h[i/2] & 0x0f
		}
		if nibble >= 8 {
			out[i] = c - 32
		}
	}
	return "0x" + string(out)
}
