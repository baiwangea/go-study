package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC20ABI is a simplified ABI for ERC20 token name function.
const ERC20ABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"}]`

func main() {
	// IMPORTANT: Replace with your own Ethereum node RPC endpoint URL.
	// You can get one for free from services like Infura or Alchemy.
	rpcEndpoint := "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID"

	// Create a new Ethereum client.
	client, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v. Please check your RPC endpoint.", err)
	}
	fmt.Println("Successfully connected to Ethereum client!")

	// Run the examples.
	queryLatestBlock(client)
	queryAccountBalance(client)
	generateNewWallet()
	// interactWithContract(client)
}

func queryLatestBlock(client *ethclient.Client) {
	fmt.Println("\n--- Querying Latest Block ---")

	// Get the latest block header.
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Latest block number:", header.Number.String())
}

func queryAccountBalance(client *ethclient.Client) {
	fmt.Println("\n--- Querying Account Balance ---")

	// Address of a well-known account (e.g., the Ethereum Foundation).
	address := common.HexToAddress("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")

	// Get the balance of the account at the latest block.
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal(err)
	}

	// The balance is in Wei. To convert it to Ether, we need to divide by 10^18.
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(1e18))

	fmt.Printf("Balance for address %s: %s ETH\n", address.Hex(), ethValue.Text('f', 4))
}

func generateNewWallet() {
	fmt.Println("\n--- Generating New Wallet ---")

	// Generate a new private key.
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Get the public key from the private key.
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	// Get the public address from the public key.
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println("Generated Wallet Address:", address)

	// Get the private key in hexadecimal format (NEVER expose this in a real application).
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Println("Generated Private Key (Hex):", common.Bytes2Hex(privateKeyBytes))
}

// This is a placeholder for a future example.
func interactWithContract(client *ethclient.Client) {
	// ... implementation to be added ...
}
