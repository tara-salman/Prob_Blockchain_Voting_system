package main

import (
	"fmt"
	"github.com/CPSSD/voting/src/blockchain"
	"github.com/CPSSD/voting/src/election"
	"log"
	"os"
	"sync"
	"strconv"
	"time"
)

var (
	tokenMsg    string = "Please enter your unique token"
	voteMsg     string = "Please enter your ballot message"
	badInputMsg string = "Unrecognised input"
	waitMsg     string = "Waiting for processes to quit"
	// to quit entirely
	quit = make(chan bool, 1)

	// to signal to stop mining
	stop = make(chan bool, 1)

	// to signal to start mining
	start = make(chan bool, 1)
	//Leader selection
	Leader string 

	// to signal to confirm stopped mining
	confirm = make(chan bool, 1)
	wg sync.WaitGroup
)
type Blockchain struct{
	chain     *blockchain.Chain
	wg sync.WaitGroup
}

func main() {
	f, err := os.OpenFile(os.Args[1]+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Can't open file")		
		panic(err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Lshortfile)
	
	
	log.Println("Setting up network config")

	var bc *Blockchain =new(Blockchain)
	bc.Init(os.Args[1],os.Args[2]) 

	fmt.Println("Welcome to voting system.")
	
	sum := 1
	for sum < 11{
		bc.Decide( strconv.FormatFloat(float64(sum)/float64(100), 'E', -1, 64), "1")		
		sum += 1
	} 	
	//bc.Decide( "0.8", "2")		
	//fmt.Printf("%v\n", waitMsg)
	//log.Printf("%v\n", waitMsg)
	//wg.Wait()
	//log.Println(bc.chain)


	bc.GetChain()
	bc.GetTally()
	
	
loop:
	/*for {
		fmt.Printf("")
	}*/
	for {
		var input string 
		fmt.Scanf("%v\n", &input)

		switch input {
		
		case "peers":
			bc.chain.PrintPeers()
		case "pool":
			bc.chain.PrintPool()
		case "chain":
			fmt.Println("Entering print chain")
			fmt.Println(bc.chain)
			fmt.Println("Exited print chain")
		case "q":
			quit <- true
			break loop
		case "b":
			fmt.Printf("Broadcasting our share of the election key\n")
			bc.chain.BroadcastShare()
			fmt.Printf("Attempting to reconstruct the election key\n")
			bc.chain.ReconstructElectionKey()
			bc.chain.PrintKey()
			ballots := bc.chain.CollectBallots()
			format := bc.chain.GetFormat()
			key := bc.chain.GetElectionKey()
			fmt.Println("Calculating the tally...")
			tally, err := format.Tally(ballots, &key)
			if err != nil {
				fmt.Println("Error calculating tally")
			}
			fmt.Println(tally)
		/*case "v":

			token = vt

			ballot = new(election.Ballot)
			vote, er =  strconv.ParseFloat(os.Args[2],64)
			vote2 = 1-vote
			err = ballot.Fill(c.GetFormat(), tokenMsg,vote, vote2)
			if er != nil {
				fmt.Println("Error with input vote")
			}
			if err != nil {
				log.Printf("Error filling out the ballot")
			} else {
				tr := c.NewTransaction(token, ballot)
				fmt.Printf("%#v \n",tr)
				go c.ReceiveTransaction(tr, nil)

			}*/
		
		default:
		}

	}
	fmt.Printf("%v\n", waitMsg)
	log.Printf("%v\n", waitMsg)
	wg.Wait()
	log.Println(bc.chain)

	
}
func (bc *Blockchain) Init (logname string,leader string) {
	var err error 
	bc.chain, err = blockchain.NewChain()
	if err != nil {
		fmt.Println("Can't do a chain ")		
		panic(err)
	}	
	filename := string(logname)
	
	bc.chain.Init(filename)

//	log.Println("Setting up network config")

	var syncDelay int = 5
	
	wg.Add(20000)
	bc.chain.Start(leader, syncDelay, quit, stop, start, confirm, &wg)
	start <- true
	//bc.chain = c
	//return bc
}
func  (bc *Blockchain) Decide(input string, eventIDs string) {
	vt := bc.chain.GetVoteToken()
	//log.Println("Hello")
	fmt.Println("Your vote token is:", vt)
	token := vt
	ballot := new(election.Ballot)
	vote, er :=  strconv.ParseFloat(input,64)
	vote2 := 1-vote
	eventID, er2 :=  strconv.ParseFloat(eventIDs,64)
	err := ballot.Fill(bc.chain.GetFormat(), tokenMsg,vote, vote2, eventID)
	if er != nil {
		fmt.Println("Error with input vote")
	}
	if er2 != nil {
		fmt.Println("Error with input eventID")
	}
	if err != nil {
		log.Println("Error filling out the ballot")
	} else {
		tr := bc.chain.NewTransaction(token, ballot)
		fmt.Printf("%#v \n",tr)
		go bc.chain.ReceiveTransaction(tr, nil)

	}

}
func (bc *Blockchain) GetChain () {
	time.Sleep(10*time.Second) 
	fmt.Println("Entering print chain")
	fmt.Println(bc.chain)
	fmt.Println("Exited print chain")
}
func (bc *Blockchain) GetTally () {
	time.Sleep(10*time.Second) 	
	ballots := bc.chain.CollectBallots()
	format := bc.chain.GetFormat()
	key := bc.chain.GetElectionKey()
	fmt.Println("Calculating the tally...")
	tally, err := format.Tally(ballots, &key)
	if err != nil {
		fmt.Println("Error calculating tally")
	}
	fmt.Println(tally)
}


