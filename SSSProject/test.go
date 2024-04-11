package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type Settings struct {
	PrivateKey         string `json:"privateKey"`
	TokenMintPublicKey string `json:"tokenMintPublicKey"`
	RPCURL             string `json:"RPCURL"`
	WSURL              string `json:"WSURL"`
}

func main() {
	// Load settings from JSON file:
	settingsFile, err := os.Open("settings.json")
	if err != nil {
		panic(err)
	}
	defer settingsFile.Close()

	var settings Settings
	err = json.NewDecoder(settingsFile).Decode(&settings)
	if err != nil {
		panic(err)
	}
	fmt.Println("Settings loaded")

	// Create a new RPC client:
	rpcClient := rpc.New(settings.RPCURL)

	// Create a new WS client (used for confirming transactions):
	wsClient, err := ws.Connect(context.Background(), settings.WSURL)
	if err != nil {
		panic(err)
	}
	fmt.Println("Rpc loaded")

	// Load account from private key:
	walletPrivateKey := solana.MustPrivateKeyFromBase58(settings.PrivateKey)
	walletPublicKey := walletPrivateKey.PublicKey()

	// The amount to mint (in the smallest unit of the token; e.g., 1 token = 1_000_000_000 lamports):
	amountToMint := uint64(100)

	// Get recent blockhash:
	recent, err := rpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}
	fmt.Println("creatin instruction")
	// Create instruction to mint tokens:
	mintToInstruction := token.NewMintToInstruction(
		amountToMint,
		solana.MustPublicKeyFromBase58(settings.TokenMintPublicKey),
		walletPublicKey,
		solana.MustPublicKeyFromBase58(settings.TokenMintPublicKey),
		nil, // No multisig signers
	)

	// Validate and build the instruction:
	mintToInstructionBuilt, err := mintToInstruction.ValidateAndBuild()
	if err != nil {
		panic(err)
	}

	// Create transaction with the mintTo instruction:
	tx, err := solana.NewTransaction(
		[]solana.Instruction{mintToInstructionBuilt},
		recent.Value.Blockhash,
	)
	if err != nil {
		panic(err)
	}
	spew.Dump(tx)

	// Sign the transaction with the wallet private key:
	// Utilisez le symbole de soulignement (_) pour ignorer la valeur retournée
	tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			// Vérifiez si la clé publique correspond à la clé publique du portefeuille
			if key.Equals(walletPublicKey) {
				// Si elle correspond, retournez la clé privée du portefeuille
				return &walletPrivateKey
			}
			return nil
		},
	)

	// Send transaction, and wait for confirmation:
	sig, err := confirm.SendAndConfirmTransaction(
		context.TODO(),
		rpcClient,
		wsClient,
		tx,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Transaction sent! Signature:", sig)
}
