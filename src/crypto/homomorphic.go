package crypto

import (
	"math/big"
	"github.com/ALTree/bigfloat"
	//"math"
	//"fmt"
)

// AddCipherTexts returns the encrypted sum of the
// individual ciphertexts. Calling this function using
// a PrivateKey will result in the function being called
// for the PublicKey.
func (key *PrivateKey) AddCipherTexts(ciphertexts ...*big.Float) (mean *big.Float, std *big.Float, ci string, err error) {
	mean, std, ci, err = key.PublicKey.AddCipherTexts(ciphertexts...)
	return
}

// AddCipherTexts accepts one or more ciphertexts
// and returns the homomorphic sums of them.
func (key *PublicKey) AddCipherTexts(ciphertexts ...*big.Float) (mean *big.Float, std *big.Float, ci string, err error) {

	if err = key.Validate(); err != nil {
		return nil, nil, "",err
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
	mean = Mean(ciphertexts)
	std = StandardDeviation(ciphertexts)
	//fmt.Printf("%0.2f \n", v)
	lower, upper := NormalConfidenceInterval(ciphertexts)
	ci = "["+lower.String()+","+upper.String()+"]"
	return mean, std, ci, nil
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
	return (new(big.Float).Quo(mean,new(big.Float).SetInt64(int64(len(nums)))))
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
