package blockchain

import (
	"encoding/hex"
	"errors"

	"github.com/blockmandu/pkg/transaction"
	"github.com/boltdb/bolt"
)

type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		cursor := b.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			txID := hex.EncodeToString(k)
			outs, err := transaction.DeserializeOutputs(v)
			if err != nil {
				return err
			}

			for outIds, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIds)
				}
			}
		}

		return nil
	})

	if err != nil {
		return 0, nil, err
	}

	return accumulated, unspentOutputs, nil
}

func (u UTXOSet) Reindex() error {
	db := u.Blockchain.DB
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
			return err
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	UTXO := u.Blockchain.FindUTXO()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txid, outs := range UTXO {
			key, err := hex.DecodeString(txid)
			if err != nil {
				return err
			}

			serialized, err := outs.Serialize()
			if err != nil {
				return err
			}

			err = b.Put(key, serialized)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (u UTXOSet) FindUTXO(PublicKey []byte) ([]transaction.TXOutput, error) {
	var UTXOs []transaction.TXOutput
	db := u.Blockchain.DB
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		cursor := b.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			outs, err := transaction.DeserializeOutputs(v)
			if err != nil {
				return err
			}

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(PublicKey) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return UTXOs, nil
}

func (u UTXOSet) Update(block *Block) error {
	db := u.Blockchain.DB

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := transaction.TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs, err := transaction.DeserializeOutputs(outsBytes)
					if err != nil {
						return err
					}

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err = b.Delete(vin.Txid)
						if err != nil {
							return err
						}
					} else {
						serialized, err := updatedOuts.Serialize()
						if err != nil {
							return err
						}

						err = b.Put(vin.Txid, serialized)
						if err != nil {
							return err
						}
					}
				}
			}

			newOutputs := transaction.TXOutputs{}
			newOutputs.Outputs = append(newOutputs.Outputs, tx.Vout...)

			serialized, err := newOutputs.Serialize()
			if err != nil {
				return err
			}

			err = b.Put(tx.ID, serialized)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
