package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

const targetBits = 24
const maxNonce = math.MaxInt64

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) ([]byte, error) {
	transactionHash, err := pow.block.HashTransaction()
	if err != nil {
		return nil, err
	}

	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			transactionHash,
			[]byte(fmt.Sprintf("%x", pow.block.Timestamp)),
			[]byte(fmt.Sprintf("%x", int64(targetBits))),
			[]byte(fmt.Sprintf("%x", int64(nonce))),
		},
		[]byte{},
	)

	return data, nil

}

func (pow *ProofOfWork) Run() (int, []byte, error) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block")
	for nonce < maxNonce {
		data, err := pow.prepareData(nonce)
		if err != nil {
			return 0, nil, err
		}

		hash = sha256.Sum256(data)
		// fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		/**
		for testing only, actual mining should be hashInt.Cmp(pow.target) == -1
		Why This Comparison?
			Proof of Work:
			The comparison ensures that the miner has done a sufficient amount of computational work.
			Finding a hash that is less than the target is computationally difficult and requires many attempts.

			Difficulty Adjustment:
			The target is adjusted based on the network's total computational power to maintain a consistent
			block generation time.

			Security:
			This process secures the network by making it computationally expensive to alter the blockchain.
			An attacker would need to redo the proof of work for all subsequent blocks to change a block's data.
		*/
		if hashInt.Cmp(pow.target) > -1 {
			break
		}

		nonce++
	}

	fmt.Print("\n\n")
	return nonce, hash[:], nil
}

func (pow *ProofOfWork) Validate() (bool, error) {
	var hashInt big.Int

	data, err := pow.prepareData(pow.block.Nonce)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1, nil
}
