package crypto

import (
	"math/big"
	"github.com/ALTree/bigfloat"
	//"math"
	//"fmt"
)
//function constains check if an int is in a list 
func contains(s [] *big.Float, e *big.Float) bool {
    for _, a := range s {
        if (a.Cmp(e)==0) {
            return true
        }
    }
    return false
}

type Events struct {
	id *big.Float  
	votes [] *big.Float
}
type EventsUnClassified struct {
	ids [] *big.Float  
	votes [] *big.Float
}
// ListEvents function take all decisions and return the list of events included 
func ListEvents ( votes [][] *big.Float) []Events {
	var ids [] *big.Float
	var events [] Events
	 
	for m,_:=range votes{ 
		if (! contains(ids, votes[m][0])) {
			ids= append(ids, votes[m][0])
			var e Events
			e.id= votes[m][0]
			e.votes= append(e.votes,votes[m][1])
			events= append(events,e)
			
		} else {
		 for i,event := range events{
		 	if (event.id.Cmp(votes[m][0])==0) {
				event.votes= append(event.votes,votes[m][1])
				events[i]=event
				
				}}
		}}
	return events
}
// AddCipherTexts returns the encrypted sum of the
// individual ciphertexts. Calling this function using
// a PrivateKey will result in the function being called
// for the PublicKey.

func (key *PrivateKey) AddCipherTexts(ciphertexts [][]*big.Float) (results [][]string,err error) {
	results, err = key.PublicKey.AddCipherTexts(ciphertexts)
	return
}
// AddCipherTexts accepts one or more ciphertexts
// and returns the homomorphic sums of them.
func (key *PublicKey) AddCipherTexts(ciphertexts [][]*big.Float) ( results [][]string,err error) {

	if err = key.Validate(); err != nil {
		return results, err
	}

	// create an encryption of voting value zero to start off
	
	//fmt.Println (total)
	//if err != nil {
	//	return nil, err
	//}

	// D(E(m1,r1).E(m2,r2) mod n^2) = m1 + m2 mod n
	/*for _, ciphertext := range ciphertexts {
		mean = new(big.Float).Add(mean, ciphertext)
		//total.Mod(total, key.NSquared)
	}
	x:=len(ciphertexts)
	//fmt.Println (x) 
	if x==0 {
		return mean, nil }
	mean= new(big.Float).Quo(mean,new(big.Float).SetInt64(int64(x)))*/
	var ids [] *big.Float
	//fmt.Printf("*********************************************")
	for a,_  := range ciphertexts {
		if (! contains(ids, ciphertexts[a][0])) {
			ids= append(ids, ciphertexts[a][0])
		}}
	var votes [][]*big.Float
	for a, _ := range ciphertexts {
			votes= append(votes,[]*big.Float{ciphertexts[a][0],ciphertexts[a][1]})
	}
	//fmt.Printf("*********************************************")
	var votesPerEvent [] Events
	votesPerEvent = ListEvents(votes)
	//fmt.Printf("*****************************%0.2f", len(ciphertexts ))
	for i, _:= range votesPerEvent {
		results =append(results, []string{"id",votesPerEvent[i].id.String(), "mean",Mean(votesPerEvent[i].votes).String(), "std",StandardDeviation(votesPerEvent[i].votes).String()})
		//std = StandardDeviation(votesPerEvent[i].votes)
		//fmt.Printf("**************************%0.2f \n", votesPerEvent[i].id)
		//lower, upper := NormalConfidenceInterval(votesPerEvent[i].votes)
		//ci = "["+lower.String()+","+upper.String()+"]"
		//eventID= votesPerEvent[i].id
	}
	//var neededTx []* Transaction
	// Then it compares the previous block transaction and append if the same id is found 
	//for _, n := range previousBlocktxs {
	//	if (contains(ids, n.EventID())) {
	//		neededTx = append (neededTx,n)
	//	}}
	//if mean == nil{ 
	//	mean=new(big.Float).SetInt64(int64(0))}
	//if std ==nil {
	//	std=new(big.Float).SetInt64(int64(0))}
	//if eventID==nil {
	//	eventID=new(big.Float).SetInt64(int64(0))}
	//if ci=="" {
	//	ci=""
	//}
	return results, nil
}
// Mean returns the mean of an integer array as a float
func Mean(nums [] *big.Float) (mean *big.Float) {
	if len(nums) == 0 {
		return new(big.Float).SetInt(big.NewInt(0))
	}

	mean = new(big.Float)
	for _, n := range nums {
		mean = new(big.Float).Add(mean,n)
	}
	return (new(big.Float).Quo(new(big.Float).Mul(mean,new(big.Float).SetInt64(int64(100))),new(big.Float).SetInt64(int64(len(nums)))))
}

func StandardDeviation(nums [] *big.Float) (dev *big.Float) {
	if len(nums) == 0 {
		return new(big.Float).SetInt(big.NewInt(0))
	}

	m := Mean(nums)
	dev = new(big.Float)
	for _, n := range nums {
	//	dev += (new(big.Float).SetInt(big.NewInt(n)) - m) * ( new(big.Float).SetInt(big.NewInt(n)) - m)
		dev= new(big.Float).Add(new(big.Float).Mul(new(big.Float).Sub( n,m), new(big.Float).Sub( n,m)),dev)
	}
	dev = new(big.Float).Quo(dev,new(big.Float).SetInt64(int64(len(nums))))
	dev = bigfloat.Pow(dev,new(big.Float).SetFloat64(0.5)) //math.Pow(dev/  big.Float(len(nums)), 0.5)
	return dev
}
func NormalConfidenceInterval(nums [] *big.Float) (lower *big.Float, upper *big.Float) {
	if len(nums) == 0 {
		return new(big.Float).SetInt(big.NewInt(0)),new(big.Float).SetInt(big.NewInt(0))
	}

	conf := 1.95996 // 95% confidence for the mean, http://bit.ly/Mm05eZ
	mean := Mean(nums)
	dev := new(big.Float).Quo(StandardDeviation(nums),bigfloat.Pow(new(big.Float).SetInt64(int64(len(nums))),new(big.Float).SetFloat64(0.5)))
	lower = new(big.Float).Sub(mean,new(big.Float).Mul(dev,new(big.Float).SetFloat64(conf)))
	upper = new(big.Float).Add(mean,new(big.Float).Mul(dev,new(big.Float).SetFloat64(conf)))
	return lower,upper
}
