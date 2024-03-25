package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = ".\\tmp\\blocks_%s"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	Lasthash []byte
	Database *badger.DB
}

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

func InitBlockChain(address string, nodeID string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeID)
	if DBexists(path) {
		fmt.Println("existing blockchain found")
		runtime.Goexit()
	}
	var lasthash []byte

	opts := badger.DefaultOptions(path)
	db, err := openDB(path, opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), genesis.Hash)

		lasthash = genesis.Hash

		return err
	})
	if err != nil {
		log.Panic(err)
	}
	blockchain := BlockChain{lasthash, db}
	return &blockchain
}

func ContinueBlockChain(nodeID string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeID)
	if !DBexists(path) {
		fmt.Println("No existing blockchain found")
		runtime.Goexit()
	}

	var lasthash []byte

	opts := badger.DefaultOptions(path)
	opts.Dir = path
	opts.ValueDir = path
	db, err := openDB(path, opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lasthash = val
			return nil
		})
		return err
	})

	if err != nil {
		log.Panic(err)
	}

	chain := BlockChain{lasthash, db}

	return &chain

}

func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}
		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}
		var lasthash []byte
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lasthash = val
			return nil
		})
		item, err = txn.Get(lasthash)
		if err != nil {
			log.Panic(err)
		}
		var lastblockdata []byte
		err = item.Value(func(val []byte) error {
			lastblockdata = val
			return nil
		})
		lastblock := Deserialize(lastblockdata)

		if block.Height > lastblock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			chain.Lasthash = block.Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func (chain *BlockChain) GetBlock(hash []byte) (Block, error) {
	var block Block
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash)
		if err != nil {
			log.Panic(err)
		}
		var blockdata []byte
		err = item.Value(func(val []byte) error {
			blockdata = val
			return nil
		})
		block = *Deserialize(blockdata)
		return err
	})
	if err != nil {
		return block, err
	}
	return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	iter := chain.Iterator()

	for {
		block := iter.Next()
		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return blocks
}

func (chain *BlockChain) GetBestHeight() int {
	var lastblock Block
	err := chain.Database.View(func(txn *badger.Txn) error {
		var lasthash []byte
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lasthash = val
			return nil
		})
		var lastblockdata []byte
		item, err = txn.Get(lasthash)
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lastblockdata = val
			return nil
		})
		lastblock = *Deserialize(lastblockdata)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastblock.Height
}

func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lasthash []byte
	var lastHeight int

	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lasthash = val
			return nil
		})
		item, err = txn.Get(lasthash)
		if err != nil {
			log.Panic(err)
		}
		var lastblockdata []byte
		err = item.Value(func(val []byte) error {
			lastblockdata = val
			return nil
		})
		lastblock := Deserialize(lastblockdata)
		lastHeight = lastblock.Height

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(transactions, lasthash, lastHeight+1)

	//put new block in database
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.Lasthash = newBlock.Hash

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, err
	}
	opts := originalOpts
	opts.Truncate = true
	db, err := badger.Open(opts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
