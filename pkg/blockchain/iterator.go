package blockchain

import "github.com/boltdb/bolt"

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{currentHash: bc.tip, db: bc.DB}
}

func (bci *BlockchainIterator) Next() *Block {
	var block *Block
	var err error

	err = bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		encodedBlock := bucket.Get(bci.currentHash)
		block, err = Deserialize(encodedBlock)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	bci.currentHash = block.PrevBlockHash
	return block

}
