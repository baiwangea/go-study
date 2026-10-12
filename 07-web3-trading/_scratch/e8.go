//go:build level

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	fmt.Println("=== 生成私钥 ===")
	priv, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	privBytes := crypto.FromECDSA(priv)
	fmt.Println("私钥长度:", len(privBytes), "字节")
	fmt.Println("私钥 hex:", hex.EncodeToString(privBytes))
	fmt.Println("priv.D (big.Int) bit len:", priv.D.BitLen())

	fmt.Println("\n=== 公钥 ===")
	pub := crypto.FromECDSAPub(&priv.PublicKey)
	fmt.Println("公钥(带 0x04 前缀) 长度:", len(pub), "首字节:", pub[0])
	fmt.Println("公钥(去前缀) 长度:", len(pub[1:]))
	fmt.Println("公钥 X bit len:", priv.PublicKey.X.BitLen(), " Y bit len:", priv.PublicKey.Y.BitLen())
	fmt.Println("压缩公钥(crypto.CompressPubkey) 长度:", len(crypto.CompressPubkey(&priv.PublicKey)))

	fmt.Println("\n=== 地址推导 ===")
	hash := crypto.Keccak256(pub[1:])
	fmt.Println("keccak256(pubkey) 长度:", len(hash), "hex:", hex.EncodeToString(hash))
	addrBytes := hash[12:]
	addr := common.BytesToAddress(addrBytes)
	fmt.Println("取后 20 字节 →", addr.Hex())
	fmt.Println("crypto.PubkeyToAddress →", crypto.PubkeyToAddress(priv.PublicKey).Hex())

	fmt.Println("\n=== 对比：sha256 是错的 ===")
	sh := sha256.Sum256(pub[1:])
	fmt.Println("sha256(pubkey) 后 20 字节 →", common.BytesToAddress(sh[12:]).Hex())
	fmt.Println("取前 20 字节(错的) →", common.BytesToAddress(hash[:20]).Hex())

	fmt.Println("\n=== EIP-55 ===")
	fmt.Println("Hex() 大小写混合:", addr.Hex())
	fmt.Println("Hex() 全小写       :", "0x"+hex.EncodeToString(addr.Bytes()))
	fmt.Println("Hex() == 全小写 hex:", addr.Hex() == "0x"+hex.EncodeToString(addr.Bytes()))
	fmt.Println("String() 同 Hex():", addr.String())
	// 大小写不同的同一个地址
	upper := common.HexToAddress("0x" + "DE0B295669A9FD93D5F28D9EC85E40F4CB697BAE")
	fmt.Println("全大写输入 → Hex():", upper.Hex())
	fmt.Println("common.HexToAddress 是否校验 checksum:", common.IsHexAddress("0xDE0B295669A9FD93D5F28D9EC85E40F4CB697BAE"))

	fmt.Println("\n=== 已知向量 ===")
	known := common.HexToAddress("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
	fmt.Println("已知地址 Hex():", known.Hex())
	fmt.Println("Bytes():", hex.EncodeToString(known.Bytes()))

	fmt.Println("\n=== HexToECDSA 往返 ===")
	back, err := crypto.HexToECDSA(hex.EncodeToString(privBytes))
	fmt.Println("往返后地址一致:", crypto.PubkeyToAddress(back.PublicKey).Hex() == addr.Hex())
	_, err = crypto.HexToECDSA("zz")
	fmt.Println("非法 hex 的错误:", err)

	fmt.Println("\n=== keystore ===")
	dir, _ := os.MkdirTemp("", "ks-demo-*")
	defer os.RemoveAll(dir)
	fmt.Println("临时目录:", dir)

	start := time.Now()
	ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)
	fmt.Println("LightScrypt 建 keyStore 耗时:", time.Since(start))

	start = time.Now()
	acct, err := ks.NewAccount("hunter2")
	if err != nil {
		panic(err)
	}
	fmt.Println("NewAccount 耗时:", time.Since(start))
	fmt.Println("account.Address:", acct.Address.Hex())
	fmt.Println("account.URL:", acct.URL.String())

	keyJSON, err := ks.Export(acct, "hunter2", "hunter2")
	if err != nil {
		panic(err)
	}
	fmt.Println("导出 JSON 长度:", len(keyJSON), "字节")
	fmt.Println("JSON 前 200 字符:", string(keyJSON[:200]))

	// 导入到一个新目录
	dir2, _ := os.MkdirTemp("", "ks-demo2-*")
	defer os.RemoveAll(dir2)
	ks2 := keystore.NewKeyStore(dir2, keystore.LightScryptN, keystore.LightScryptP)
	acct2, err := ks2.Import(keyJSON, "hunter2", "hunter2")
	if err != nil {
		panic(err)
	}
	fmt.Println("导入后地址:", acct2.Address.Hex(), "一致:", acct2.Address == acct.Address)

	fmt.Println("\n=== 解锁签名 ===")
	if err := ks.Unlock(acct, "hunter2"); err != nil {
		panic(err)
	}
	defer ks.Lock(acct.Address)
	msg := []byte("hello web3")
	sig, err := ks.SignHash(acct, crypto.Keccak256(msg))
	if err != nil {
		panic(err)
	}
	fmt.Println("签名长度:", len(sig))
	recovered, err := crypto.SigToPub(crypto.Keccak256(msg), sig)
	if err != nil {
		panic(err)
	}
	fmt.Println("恢复出的地址:", crypto.PubkeyToAddress(*recovered).Hex())
	fmt.Println("和原地址一致:", crypto.PubkeyToAddress(*recovered) == acct.Address)

	fmt.Println("\n=== 错误密码 ===")
	_, err = ks.Export(acct, "wrong", "x")
	fmt.Println("错密码导出错误:", err)
}
