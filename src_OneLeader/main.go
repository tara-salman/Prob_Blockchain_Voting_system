package main

import (
	"fmt"
	"github.com/CPSSD/voting/src/blockchain"
	"github.com/CPSSD/voting/src/election"
	"log"
	"os"
	"sync"
	"strconv"
)

var (
	tokenMsg    string = "Please enter your unique token"
	voteMsg     string = "Please enter your ballot message"
	badInputMsg string = "Unrecognised input"
	waitMsg     string = "Waiting for processes to quit"
)

func main() {
	
	f, err := os.OpenFile(os.Args[1]+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Can't open file")		
		panic(err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Lshortfile)

	c, err := blockchain.NewChain()
	if err != nil {
		fmt.Println("Can't do a chain ")		
		panic(err)
	}
	log.Println("Setting up network config")

	filename := string(os.Args[1])

	c.Init(filename)

	// to quit entirely
	quit := make(chan bool, 1)

	// to signal to stop mining
	stop := make(chan bool, 1)

	// to signal to start mining
	start := make(chan bool, 1)

	// to signal to confirm stopped mining
	confirm := make(chan bool, 1)

	var syncDelay int = 10
	var wg sync.WaitGroup
	wg.Add(4)
	c.Start(syncDelay, quit, stop, start, confirm, &wg)
	start <- true

	fmt.Println("Welcome to voting system.")
	vt := c.GetVoteToken()
	fmt.Println("Your vote token is:", vt)
	token := vt
	ballot := new(election.Ballot)
	vote, er :=  strconv.ParseFloat(os.Args[2],64)
	vote2 := 1-vote
	eventID, er2 :=  strconv.ParseFloat(os.Args[3],64)
	err = ballot.Fill(c.GetFormat(), tokenMsg,vote, vote2, eventID)
	if er != nil {
		fmt.Println("Error with input vote")
	}
	if er2 != nil {
		fmt.Println("Error with input eventID")
	}
	if err != nil {
		log.Println("Error filling out the ballot")
	} else {
		tr := c.NewTransaction(token, ballot)
		fmt.Printf("%#v \n",tr)
		go c.ReceiveTransaction(tr, nil)

	}

loop:
	/*for {
		fmt.Printf("")
	}*/
	for {
		var input string
		fmt.Scanf("%v\n", &input)

		switch input {
		
		case "peers":
			c.PrintPeers()
		case "pool":
			c.PrintPool()
		case "chain":
			fmt.Println("Entering print chain")
			fmt.Println(c)
			fmt.Println("Exited print chain")
		case "q":
			quit <- true
			break loop
		case "b":
			fmt.Printf("Broadcasting our share of the election key\n")
			c.BroadcastShare()
			fmt.Printf("Attempting to reconstruct the election key\n")
			c.ReconstructElectionKey()
			c.PrintKey()
			ballots := c.CollectBallots()
			format := c.GetFormat()
			key := c.GetElectionKey()
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
	log.Println(c)
}
