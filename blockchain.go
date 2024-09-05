package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The first ever coinbase trainsaction on Blockmandu"
)

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func dbExists() bool {
	_, err := os.Stat(dbFile)
	return !os.IsNotExist(err)
}

func NewBlockchain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesisBlock := NewGenesisBlock(cbtx)

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
		log.Panic(err)
	}

	return &Blockchain{tip: tip, db: db}
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := map[string][]int{}
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if _, ok := spentTXOs[txID]; ok {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
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

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulate := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulate < amount {
				accumulate += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulate >= amount {
					break Work
				}
			}
		}
	}

	return accumulate, unspentOutputs
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
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
	if err != nil {
		log.Panic(err)
	}
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{currentHash: bc.tip, db: bc.db}
}

func (bci *BlockchainIterator) Next() *Block {
	var block *Block

	bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		encodedBlock := bucket.Get(bci.currentHash)
		block = Deserialize(encodedBlock)

		return nil
	})

	bci.currentHash = block.PrevBlockHash
	return block

}
