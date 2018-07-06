package election

import (
	"errors"
	"fmt"
	"github.com/CPSSD/voting/src/crypto"
	"math/big"
	"strings"
)

var (
	InvalidFormatError = errors.New("Invalid format was supplied; bad number of selections.")
)

// Ballot contains the structure of a ballot which a given
// user will fill out. It contains the VoteToken of the
// voter, along with the selections made by the voter.
type Ballot struct {
	VoteToken     string      // VT of the voter who owns ballot
	NumSelections int         // number of selections in the ballot
	Selections    []Selection // list of selections on the ballot
}

// Selection contains details on particular selection by
// a user on a ballot, including its name and vote value.
type Selection struct {
	Name  string   // name of this selection option
	Vote  *big.Float // value should be encrypted with PrivKE
	Proof []byte   // value used as the zero-knowledge proof
}

// Format defines the format of a ballot, and should be used
// to ensure that ballots follow the format defined for a
// vote.
type Format struct {
	NumSelections int
	Selections    []Selection
}

// Fill uses the defined Format f and the VoteToken vt and
// takes a user through the prompts to fill out a ballot.
// The ballot is returned in the value of b.
func (b *Ballot) Fill(f Format, vt string, input float64, input2 float64) (err error) {
	if len(f.Selections) != f.NumSelections {
		return InvalidFormatError
	}

	b.VoteToken = vt
	b.NumSelections = f.NumSelections
	b.Selections = make([]Selection, f.NumSelections)
	vote := big.NewFloat(float64(input))
	selection := Selection{
			Name:  "Attack",
			Vote:  vote,
			Proof: make([]byte, 0),
		}
	b.Selections[0] = selection
	vote = big.NewFloat(float64(input2))
	selection = Selection{
			Name:  "Normal",
			Vote: vote ,
			Proof: make([]byte, 0),
		}
	b.Selections[1] = selection
	/*for i, s := range f.Selections {
		//fmt.Printf("Enter your selection (between 0 and 1) for Candidate %v: ", s.Name)
		//var input float64
		//fmt.Scanf("%v\n", &input)
		//if input > 1 {
		//	fmt.Println("Wrong input please use something between 0 and 1") 
		//	b.Fill (f, vt) 
		//}
		vote := big.NewFloat(float64(input))
		//fmt.Printf("Vote: %v Input: %v \n", vote, input)
		selection := Selection{
			Name:  s.Name,
			Vote:  vote,
			Proof: make([]byte, 0),
		}

		b.Selections[i] = selection
		//fmt.Printf ("%v\n",b.Selections[i])
	}*/

	// TODO: let user review inputs before returning

	return nil
}

// CreateFormat allows for a defined Format to be created
// for an election, and takes a user through defining the
// selections available on a ballot.
func CreateFormat() (f *Format) {
	fmt.Printf("How many selections? ")
	var input int
	fmt.Scanf("%v\n", &input)

	f = &Format{
		NumSelections: input,
		Selections:    make([]Selection, input),
	}

	fmt.Println("Use double quotes for description entries")
	for i := 0; i < input; i++ {
		fmt.Printf("Enter user description for selection %v: ", i+1)
		var desc string

		fmt.Scanf("%q\n", &desc)
		fmt.Println("You entered:", desc)

		desc = strings.Trim(desc, " \n")

		s := Selection{
			Name: desc,
		}

		f.Selections[i] = s
	}

	return f
}

// Tally represents a map of selection names to their
// total counts.
type Tally struct {
	Totals map[string]*big.Float
	STD map[string]*big.Float
	CI map[string] string
}

// String representation of a Tally.
func (t Tally) String() (str string) {
	for name, result := range t.Totals {
		str = str + name + ": " + result.String() + " percent\t"
		str = str + "STD : " + t.STD[name].String() + "\t"
		str = str+ "CI:" + t.CI[name]+"\n"
	}
	return str
}

// Tally creates a Tally of the Ballots in bs, according
// to the Format in f. The final results are decrypted using
// the PrivateKey provided.
func (f *Format) Tally(bs *[]Ballot, key *crypto.PrivateKey) (t *Tally, err error) {

	t = &Tally{
		Totals: make(map[string]*big.Float, 0),
		STD: make(map[string]*big.Float, 0),
		CI: make(map[string]string, 0),
	}

	selectionCounts := make(map[string][]*big.Float, 0)

	for _, s := range f.Selections {
		selectionCounts[s.Name] = make([]*big.Float, len(*bs))
	}

	for i, b := range *bs {
		for _, s := range b.Selections {
			if _, ok := selectionCounts[s.Name]; ok {
				selectionCounts[s.Name][i] = s.Vote
				//fmt.Printf("%#v",selectionCounts[s.Name][i])
			}
		}
	}

	// TODO: decrypt each sub tally
	for name, count := range selectionCounts {		
		mean,std,ci, err := key.AddCipherTexts(count...)
		if err != nil {
			return t, err
		}
		/*mean,std,ci, err := key.Decrypt(sum)
		if err != nil {
			return t, err
		}*/
		t.Totals[name] = mean
		t.STD[name]=std
		t.CI[name]= ci
	}
	return t, err
}

func (b Ballot) String() (str string) {
	
	str = str+ "\n "
	for _, s := range b.Selections {
		str = str + " // Selection:    " + s.Name
		str = str + "        Vote:    " + fmt.Sprint(s.Vote)
		str = str+ "\n \t\t"
	}

	return str
}
