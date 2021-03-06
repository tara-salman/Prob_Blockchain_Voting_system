package blockchain

import (
	"github.com/CPSSD/voting/src/crypto"
	"log"
	"strconv"
	"sync"
	"time"
	"fmt"
	"math/rand"
	//"math/big"
)

// Chain contains all of the information required for the system
// to function.
type Chain struct {
	Peers               chan map[string]bool
	TransactionPool     chan []Transaction
	TransactionsReady   chan []Transaction
	CurrentTransactions chan []Transaction
	BlockUpdate         chan BlockUpdate
	KeyShares           chan map[string]ElectionSecret
	SeenTrs             chan map[string]bool
	head                *Block
	blocks              chan []Block
	tmpblocks	    chan []Block
	conf                Configuration
	Tally		    string
}

func NewChain() (c *Chain, err error) {
	c = &Chain{
		Peers:               make(chan map[string]bool, 1),
		TransactionPool:     make(chan []Transaction, 1),
		TransactionsReady:   make(chan []Transaction, 1),
		CurrentTransactions: make(chan []Transaction, 1),
		BlockUpdate:         make(chan BlockUpdate, 1),
		KeyShares:           make(chan map[string]ElectionSecret, 1),
		SeenTrs:             make(chan map[string]bool, 1),
		head:                NewBlock(c),
		blocks:              make(chan []Block, 1),
		//Tally:		     make(chan big.Float, 0.0),
	}
	pool := make([]Transaction, 0)
	c.TransactionPool <- pool
	seenTrs := make(map[string]bool, 0)
	c.SeenTrs <- seenTrs
	keyShares := make(map[string]ElectionSecret, 0)
	c.KeyShares <- keyShares
	blocks := make([]Block, 0)
	c.Tally = " Not Calculated yet"
	c.blocks <- blocks
	rand.Seed(time.Now().UTC().UnixNano())			
	return c, nil
}

// ReconstructElectionKey will attempt to reconstruct the
// election key from the shares currently available to a node.
func (c *Chain) ReconstructElectionKey() {
	shares := <-c.KeyShares
	c.KeyShares <- shares

	lambdaShares := make([]crypto.Share, len(shares))
	muShares := make([]crypto.Share, len(shares))

	var i int
	for _, s := range shares {
		lambdaShares[i] = s.Lambda
		muShares[i] = s.Mu
		i++
	}

	lambda, err := crypto.Interpolate(lambdaShares, c.conf.ElectionLambdaModulus)
	if err != nil {
		log.Println("Error reconstructing the lambda value for the election key")
		log.Fatalln(err)
	}

	mu, err := crypto.Interpolate(muShares, c.conf.ElectionMuModulus)
	if err != nil {
		log.Println("Error reconstructing the mu value for the election key")
		log.Fatalln(err)
	}

	c.conf.ElectionKey.Lambda = lambda
	c.conf.ElectionKey.Mu = mu
}

// String representation of a Chain.
func (c Chain) String() (str string) {
	blocks := <-c.blocks
	c.blocks <- blocks
	for i, b := range blocks {
		str = str + "Block " + strconv.Itoa(i) + ": \b" + b.String() + "\n"
	}
	str= str+ " \nTally : "+ fmt.Sprint(c.Tally)+ "\n"
	return "Chain:\n " + str
}

// schedulePeerSync will regularly sync peer lists with its known
// peers.
func (c *Chain) schedulePeerSync(syncDelay int, quit chan bool, wg *sync.WaitGroup) {
	timer := time.NewTimer(time.Second)
loop:
	for {
		select {
		case <-quit:
			log.Println("Peer syncing process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop
		case <-timer.C:
			log.Println("About to sync peers")
			c.syncPeers()
			timer = time.NewTimer(time.Second * time.Duration(syncDelay))
		}
	}
}

// removeSeenTransactions will return an array of transactions which do not
// occur in the map of seen transaction tokens
func (c *Chain) removeSeenTransactions(trs []Transaction, seen map[string]bool) (out []Transaction) {

	for _, tr := range trs {
		if _, ok := seen[tr.Header.Signature.R.String()]; !ok {
			if valid := c.ValidateSignature(&tr); valid {
				out = append(out, tr)
				seen[tr.Header.Signature.R.String()] = true
			}
		}
	}
	log.Println("Hello from chain")
	return out
}

// scheduleMining is responsible for the logic of creating new
// blocks in the chain.
func (c *Chain) scheduleMining(leader string, quit, stopMining, startMining, confirmStopped chan bool, wg *sync.WaitGroup) {
	timer := time.NewTimer(time.Second)
start:

	log.Println("Waiting for the signal to start mining")
	//fmt.Println(startMining)
	_ = <-startMining
	//fmt.Println("Here") 
	log.Println("Got the signal, about to start mining")

loop:
	
	for {
		//fmt.Println("Start")
		select {

		default:
			// By default, we wait for timer to expire, then we will check
			// to see if there are enough transactions in the pool that we
			// can create a block from.
			_ = <-timer.C

			// Get the pool and see if it is longer than the constant blockSize
			pool := <-c.TransactionPool
			//fmt.Println(len(pool))
			if len(pool) >= blockSize {
				// if so, we will put blockSize worth of transactions into
				// the TransactionsReady channel, and replace the rest of the
				// transactions

				c.TransactionsReady <- pool[:blockSize]
				c.TransactionPool <- pool[blockSize:]
			} else {
				c.TransactionPool <- pool
			}
			// Reset the timer
			timer = time.NewTimer(time.Second * time.Duration(hashingDelay))

		case <-quit:
			log.Println("Mining process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop

		case <-stopMining:
			log.Println("Mining process received signal to stop activities")
			//fmt.Println("Mining process received signal to stop activities")
			c.CurrentTransactions <- make([]Transaction, 0)
			confirmStopped <- true
			goto start

		case blockPool := <-c.TransactionsReady:
			log.Println("We have enough transactions to create a block")
			//fmt.Println("We have enough transactions to create a block")
			// make a backup in case we need to stop mining
			tmpTrs := blockPool

			for _, tr := range blockPool {
				// signatures have been verified before being added to the pool
				c.head.addTransaction(&tr, c)
			}

			blocks := <-c.blocks
			c.blocks <- blocks
			ballots := c.CollectBallots()
			//fmt.Println(ballots)
			format := c.GetFormat()
			key := c.GetElectionKey()
			
			tally, err := format.Tally(ballots, &key)
			if err != nil {
				fmt.Println("Error calculating tally")
			}
			c.Tally =tally.String()
			
			fmt.Println("Create:",time.Now(), c.GetVoteToken())

			if len(blocks) != 0 {
				c.head.Header.ParentHash = blocks[len(blocks)-1].Proof
			} else {
				c.head.Header.ParentHash = *new([32]byte)
			}
			
			// We choose our block or other people block based on the the slake in the block
			
			// compute block hash until created or stopped by new longest chain

			//leader := make (chan int,5)			
			//	leaderNo := rand.Intn(1)
			//leader <-leaderNo
			//fmt.Println(leaderNo)
			stop:=true 
			/*if (c.GetVoteToken()==leader){
			*/
			//time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			stop = c.head.CalculateHash(stopMining)
			//	stop =false
			//}
			/*if (leaderNo==1 && c.GetVoteToken()=="Bp"){
				stop = c.head.CalculateHash(stopMining)
				
			//	stop=false
			}
			if (leaderNo==2 && c.GetVoteToken()=="c2"){
				stop= c.head.CalculateHash(stopMining)
			//	stop=false
			} 
			if (leaderNo==3 && c.GetVoteToken()=="Ds"){
				stop = c.head.CalculateHash(stopMining)
			//	stop=false
			}*/
			
			
				
				
						/*} else {
							time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
						}
					}
				}
			}*/
			//fmt.Println(stop) */
			if stop {
				
				log.Println("Mining process received signal to stop activities")
				//tmpTrs2 := <-c.CurrentTransactions
				//fmt.Println(tmpTrs)
				// notify what transactions we were working with
				//if tmpTrs2 !=tmpTrs {
				c.CurrentTransactions <- tmpTrs
				//}
				//fmt.Println("Mining process received signal to stop activities")
				c.head = NewBlock(c)
				confirmStopped <- true
				//fmt.Println("hello") 
				goto start
			} else {
				
				log.Println("Mining process created a block")
				fmt.Println("I created a block")
				seenTrs := <-c.SeenTrs
				for _, tr := range c.head.Transactions {
						seenTrs[tr.Header.Signature.R.String()] = true
				}
				c.SeenTrs <- seenTrs

				blocks := <-c.blocks
				c.blocks <- append(blocks, *c.head)

				bl := *c.head
				c.head = NewBlock(c)
				ballots := c.CollectBallots()
				//fmt.Println(ballots)
				format := c.GetFormat()
				key := c.GetElectionKey()
				//fmt.Println("Calculating the tally...")
				tally, err := format.Tally(ballots, &key)
				if err != nil {
					fmt.Println("Error calculating tally")
					fmt.Print("********************")
					fmt.Println(err)
				}
				c.Tally =tally.String()
					
				go c.sendBlock(&bl)
			}
		}
	}
}

// Start will begin some of the background routines required for the running
// of the blockchain such as searching for new peers, mining blocks, handling
// chain updates, and listening for new key shares.
func (c *Chain) Start(leader string, delay int, quit, stop, start, confirm chan bool, w *sync.WaitGroup) {

	// check for new peers every "delay" seconds
	log.Println("Starting peer syncing process...")
	go c.schedulePeerSync(delay, quit, w)

	// be processing transactions aka making blocks
	log.Println("Starting mining process...")
	go c.scheduleMining(leader, quit, stop, start, confirm, w)

	// be ready to process new blocks and consensus forming
	log.Println("Starting chain management process...")
	go c.scheduleChainUpdates(quit, stop, start, confirm, w)

	// be listening for new shares
	log.Println("Starting key share collection process...")
	go c.scheduleKeyShareBroadcasting(delay, quit, w)
}

// scheduleKeyShareBroadcasting will regularly broadcast the
// list of currently known key shares to the network.
func (c *Chain) scheduleKeyShareBroadcasting(delay int, quit chan bool, wg *sync.WaitGroup) {
	timer := time.NewTimer(time.Second * time.Duration(delay))
loop:
	for {
		select {
		case <-quit:
			log.Println("Key share collection process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop

		case <-timer.C:
			log.Println("About to broadcast key shares")
			c.broadcastKeyShares()
			timer = time.NewTimer(time.Second * time.Duration(delay))

		}
	}
}

// scheduleChainUpdates will handle new block updates, and communicates
// with the mining process to help manage when and what to mine.
func (c *Chain) scheduleChainUpdates(quit, stopMining, startMining, confirmStopped chan bool, wg *sync.WaitGroup) {
loop:
	for {
		select {
		case <-quit:
			log.Println("Chain update process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop

		case blu := <-c.BlockUpdate:
			log.Println("Handling block update")

			blocks := <-c.blocks
			c.blocks <- blocks
			newBlocks := append(blocks, blu.LatestBlock)

			// validate the proposed new chain
			valid, seen := c.validate(&newBlocks)

			if valid {

				log.Println("Update contains valid next block")

			} else if !valid && blu.ChainLength > uint32(len(blocks)) {

				log.Println("Possible new longer chain;", blu.ChainLength, "vs", uint32(len(blocks)))
				log.Println("Getting alt chain")
				altChain, err := c.getChainUpdateFrom(blu.Peer)
				if err != nil {
					log.Println("There was a problem getting the alt chain")
					continue
				}

				// make sure it is longer
				if len(*altChain) < len(blocks) {
					log.Println("Alt chain is shorter")
					continue
				}

				// validate the new chain
				newBlocks = *altChain
				ballots := c.CollectBallots()
				//fmt.Println(ballots)
				format := c.GetFormat()
				key := c.GetElectionKey()
				//fmt.Println("Calculating the tally...")
				tally, err := format.Tally(ballots, &key)
				if err != nil {
					fmt.Println("Error calculating tally")
				}
				c.Tally =tally.String()

				valid, seen = c.validate(altChain)
				if valid {
					log.Println("Alt chain is valid")
				}
			}

			// if newBlocks is a valid chain...
			if valid {

				log.Println("Sending signal to stop mining")
								
				stopMining <- true

				_ = <-confirmStopped
				log.Println("We have stopped mining")
				//fmt.Println("Sending signal to stop mining")
				// set the new chain of blocks
				oldBlocks := <-c.blocks
				c.blocks <- newBlocks

				
				// set the new map of seen transactions
				_ = <-c.SeenTrs
				c.SeenTrs <- seen
				// set the new pool of transactions still to be mined
				oldPool := <-c.TransactionPool
				currentTrs := <-c.CurrentTransactions

				oldChainTrs := extractTransactions(&oldBlocks)

				allTrs := append(oldPool, currentTrs...)
				allTrs = append(allTrs, *oldChainTrs...)

				newPool := c.removeSeenTransactions(allTrs, seen)
				ballots := c.CollectBallots()
				fmt.Println(c.blocks)
				fmt.Println("Recieved:",time.Now(),c.GetVoteToken())
				format := c.GetFormat()
				key := c.GetElectionKey()
				//fmt.Println("Calculating the tally...")
				tally, err := format.Tally(ballots, &key)
				if err != nil {
					fmt.Println("Error calculating tally")
				}
				c.Tally =tally.String()

               	 		go c.broadcastOldTransactions(&newPool)

				c.TransactionPool <- newPool

				go c.sendBlock(&blu.LatestBlock)

				log.Println("Sending signal to start mining again")
				startMining <- true
			} else {
				log.Println("Alt chain was not valid")
				//time.Sleep(1000 * time.Microsecond)
				//startMining <- true
			}
		}
	}
}

func (c *Chain) broadcastOldTransactions(trs *[]Transaction) {
    log.Println("Broadcasting old transactions")

    for _, tr := range *trs {
        c.SendTransaction(&tr)
    }

    log.Println("Done broadcasting old transactions")
}

// validate will validate a set of blocks and their transactions.
func (c *Chain) validate(blocks *[]Block) (valid bool, seen map[string]bool) {

	seen = make(map[string]bool, 0)
	parent := *new([32]byte)

	for _, bl := range *blocks {

		// validate the transactions in the block
		for _, tr := range bl.Transactions {
			if valid := c.ValidateSignature(&tr); !valid {
				log.Println("Invalid chain - badly signed transaction:", tr.Header.VoteToken)
				return false, seen
			}
			if _, ok := seen[tr.Header.Signature.R.String()]; ok {
				log.Println("Invalid chain - duplicated transactions:", tr.Header.VoteToken)
				return false, seen
			}
			//fmt.Println(seen)
			seen[tr.Header.Signature.R.String()] = true
		}

		valid, hash := bl.validate(parent)

		if !valid {
			log.Println("Invalid chain - bad hash of block to parent")
			return false, seen
		}
		parent = hash
	}
	return true, seen
}

// TODO: check the chain in reverse order ie. most
// recent blocks first: hypothesis is that if a
// transaction has been seen before, it will be
// seen more recently.
func (c *Chain) contains(t *Transaction) bool {
	blocks := <-c.blocks
	for _, b := range blocks {
		if b.contains(t) {
			c.blocks <- blocks
			return true
		}
	}
	c.blocks <- blocks
	return false
}
