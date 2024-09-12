package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	common "github.com/blockmandu/pkg/commons"
	"github.com/blockmandu/pkg/transaction"
	"github.com/blockmandu/pkg/wallet"
	"github.com/boltdb/bolt"
)

const (
	dbFile              = "./resources/blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The first ever coinbase trainsaction on Blockmandu"
)

type Blockchain struct {
	DB  *bolt.DB
	tip []byte
}

func dbExists() bool {
	_, err := os.Stat(dbFile)
	return !os.IsNotExist(err)
}

func NewBlockchain(address string) (*Blockchain, error) {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		return nil, err
	}

	bc := Blockchain{tip: tip, DB: db}

	return &bc, nil
}

func CreateBlockchain(address string) (*Blockchain, error) {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	cbtx, err := transaction.NewCoinbaseTX(address, genesisCoinbaseData)
	if err != nil {
		return nil, err
	}

	genesisBlock := NewGenesisBlock(cbtx)

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("l"), genesisBlock.Hash)
		if err != nil {
			return err
		}

		tip = genesisBlock.Hash
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Blockchain{tip: tip, DB: db}, nil
}

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []transaction.Transaction {
	var unspentTXs []transaction.Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []transaction.TXOutput {
	var UTXOs []transaction.TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

func (bc *Blockchain) MineBlock(txs []*transaction.Transaction) error {
	var lastHash []byte

	for _, tx := range txs {
		verified, err := bc.VerifyTransaction(tx)
		if err != nil {
			return err
		}

		if !verified {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		return err
	}

	newBlock := NewBlock(txs, lastHash)

	return bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash

		return nil
	})
}

func (bc *Blockchain) FindTransaction(ID []byte) (transaction.Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {

			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("Transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *transaction.Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return err
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) (bool, error) {
	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return false, nil
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) (*transaction.Transaction, error) {
	var inputs []transaction.TXInput
	var outputs []transaction.TXOutput

	wallets, err := wallet.NewWallets()
	if err != nil {
		return nil, err
	}

	wallet := wallets.GetWallet(from)
	pubKeyHash := common.HashPubKey(wallet.PublicKey)

	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		return nil, fmt.Errorf("ERROR: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, err
		}

		for _, out := range outs {
			input := transaction.TXInput{Txid: txID, Vout: out, Signature: nil, PubKey: wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *transaction.NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *transaction.NewTXOutput(acc-amount, from))
	}

	tx := transaction.Transaction{ID: nil, Vin: inputs, Vout: outputs}
	id, err := tx.Hash()
	if err != nil {
		return nil, err
	}

	tx.ID = id
	bc.SignTransaction(&tx, wallet.PrivateKey)

	return &tx, nil
}
