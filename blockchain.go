package main

import (
	"github.com/boltdb/bolt"
)

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
)

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func NewBlockChain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		panic("unable to serialize block" + err.Error())
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			genesis := NewBlock("Genesis Block!", []byte{})
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}

			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				return err
			}
		} else {
			tip = bucket.Get([]byte("l"))
		}

		return nil
	})
	if err != nil {
		panic("unable initialize blockchain" + err.Error())
	}

	return &Blockchain{tip: tip, db: db}
}

func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		lastHash = bucket.Get([]byte("l"))
		return nil
	})
	if err != nil {
		panic("unable to get lastHash" + err.Error())
	}

	newBlock := NewBlock(data, lastHash)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		err = bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		bucket.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		panic("unable to add new block" + err.Error())
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
