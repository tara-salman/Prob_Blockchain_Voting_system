package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	//"reflect"
	"math/rand"
	"strconv"
	//"strings"
	"time"
	"fmt"
	//"os"
	"github.com/CPSSD/voting/src/election"
)

// Block contains a set of transactions, a proof of work, and
// a header with additional information.
type Block struct {
	Transactions []Transaction
	Header       BlockHeader
	Proof        [32]byte
	Tally        string
	Stake 	     int
	Creator	     string
}

// BlockHeader contains the hash of the block's transactions,
// the hash of its parent block, a timestamp and the nonce used
// in the creation of the proof of work.
type BlockHeader struct {
	MerkleHash [32]byte
	ParentHash [32]byte
	Timestamp  uint32
	Nonce      uint32
}
//Collect Block Ballots
func (b *Block) CollectBallots() *[]election.Ballot {
	//log.Println("Gathering ballots from the chain")
	//blocks := <-c.blocks
	//c.blocks <- blocks

	ballots := make([]election.Ballot, 0)

	//for _, bl := range blocks {
	for _, tr := range b.Transactions {
		ballots = append(ballots, tr.Ballot)
	}
	//}
	//log.Println("Collected the ballots from the chain")
	//fmt.Printf("%#v",&ballots)
	return &ballots
}

// NewBlock returns an empty initalized block.
func NewBlock(c *Chain) (b *Block) {
	b = &Block{
		Transactions: make([]Transaction, 0, blockSize),
	}
	if c==nil{
		fmt.Println("Null chain")
		b.Tally =""
		return b
	}
	ballots := b.CollectBallots()
	format := c.GetFormat()
	//fmt.Println("Error calculating tally")
	key := c.GetElectionKey()
	tally, err := format.Tally(ballots, &key)
	if err != nil {
		fmt.Println("Error calculating tally")
		fmt.Print("********************")
		fmt.Println(err)
	}
	b.Tally =tally.String()
	b.Creator= c.GetVoteToken()
	return b
}

// String representation of a block
func (b Block) String() (str string) {
	//str = str + "\n // Time:          " + fmt.Sprint(b.Header.Timestamp)
	str = str + "\n // Proof of Work: " + hex.EncodeToString(b.Proof[:15]) + "..."
	//str = str + "\n // Merkle Hash:   " + hex.EncodeToString(b.Header.MerkleHash[:])
	str = str + "\n // Parent Proof:  " + hex.EncodeToString(b.Header.ParentHash[:15])
	//str = str + "\n // Nonce:         " + fmt.Sprint(b.Header.Nonce)
	str = str + "\n\n"
	str = str + "\n // Tally:  " + b.Tally
	str = str + "\n"
	for i, t := range b.Transactions {
		str = str + "Transaction " + strconv.Itoa(i) + ": " + t.String() + "\n"
	}
	str = str + "\n // Creator:  " + b.Creator
	return str
}

// addTransaction will add a transaction to a block.
func (b *Block) addTransaction(t *Transaction, c *Chain) (isFull bool) {
	log.Println("Adding transaction")
	
	b.Transactions = append(b.Transactions, *t)
	if c==nil{
	//	fmt.Println("Null chain")
		b.Tally =""
		return len(b.Transactions) == cap(b.Transactions)
	}
	ballots := b.CollectBallots()
	format := c.GetFormat()
	//fmt.Println("Error calculating tally")
	key := c.GetElectionKey()
	tally, err := format.Tally(ballots, &key)
	if err != nil {
		fmt.Println("Error calculating tally")
		fmt.Print("********************")
		fmt.Println(err)
	}
	b.Tally =tally.String()
	return len(b.Transactions) == cap(b.Transactions)
}

// isBlockValid makes sure block is valid by checking index
// and comparing the hash of the previous block
func (bl *Block) validate(parent [32]byte) (isValid bool, hash [32]byte) {
	merkle := merkleHash(bl.Transactions)
	tmpBl := &Block{
		Header: BlockHeader{
			MerkleHash: merkle,
			ParentHash: parent,
			Timestamp:  bl.Header.Timestamp,
			Nonce:      bl.Header.Nonce,
		},
	}

	var data []byte
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(&tmpBl.Header)
	if err != nil {
		return false, hash
	}

	data = append(merkle[:], buf.Bytes()...)
	log.Println("Recreated data is: ")
	log.Println(hex.EncodeToString(data))
	hash = sha256.Sum256(data)
	log.Println("Recreated hash is: ")
	log.Println("Hello")
	log.Println(hex.EncodeToString(bl.Proof[:]))
	log.Println(hex.EncodeToString(hash[:]))
	if hash != bl.Proof{
		return false, hash
	}
	
	return true, hash
}

// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func (b *Block) CalculateHash(stop chan bool) ( stopped bool) {
	//start := time.Now()
	merkle := b.getMerkleHash()
	altB := b
	//prefix := strings.Repeat("0", prefixLen)

	b.Header.Timestamp = uint32(time.Now().Unix())
	data := make([]byte, 0)
	hash := *new([32]byte)
	select {	
		case <-stop:
			//log.Println("Interrupting POW after", time.Since(start))
				//defer f.Close()
			return true
		default:
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			err := enc.Encode(&altB.Header)
			if err != nil {
				log.Fatalln(err)
			}
			data = append(merkle[:], buf.Bytes()...)
			hash = sha256.Sum256(data)
			b.Proof=hash
	//bl.Header.Timestamp = uint32(time.Now().Unix())
			randomInt := rand.Intn(100)
			b.Stake= randomInt*len(b.Transactions)
	}	
	return false
	
}

// MerkleHash will get the hash of the transactions in a block.
func (b *Block) getMerkleHash() (hash []byte) {
	h := merkleHash(b.Transactions)
	b.Header.MerkleHash = h
	return h[:]
}

// merkleHash will get the hash of a slice of transactions.
func merkleHash(trs []Transaction) (hash [32]byte) {
	l := len(trs)
	if l == 1 {
		return trs[0].Header.BallotHash
	}
	hl := merkleHash(trs[:l/2])
	hr := merkleHash(trs[l/2:])
	return sha256.Sum256(append(hl[:], hr[:]...))
}

func (b *Block) contains(t *Transaction) bool {
	for _, tr := range b.Transactions {
		if (t.Header.Signature.R.Cmp(tr.Header.Signature.R)==0) {
			return true
		}
	}
	return false
}


// extractTransactions will gather all the transactions in a
// slice of blocks.
func extractTransactions(blocks *[]Block) *[]Transaction {
	trs := make([]Transaction, len(*blocks)*blockSize)
	var i int
	for _, bl := range *blocks {
		i += copy(trs[i:], bl.Transactions)
	}
	return &trs
}

// checkProof will check for a partial hash collision.
func checkProof(prefix string, len int, hash [32]byte) bool {
	s := hex.EncodeToString(hash[:])
	return s[:len] == prefix
}

