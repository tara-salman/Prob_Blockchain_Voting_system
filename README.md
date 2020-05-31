This source code is extending a blockchain voting system to probabilistic blockchain. Original code https://github.com/CPSSD/voting 

doc file has the original code documentation 

src file is the orignal source changed to probabilistic blockchain.

src_PoW file is for proof of work mining algorithm

src_OneLeader is for determined leader 

src_RandomMiner is for random leader (a concept that breaks one leader)

paper is for probabilstic blockchain paper

for src, src_PoW, src_OneLeader, and src_RandomMiner:

src/blockchain contain the main blockchain entinties including blocks, transactions, and chain. Our major changes there is in block.go and transaction.go 

src/cryto is encryption/decreption purposes of votes. We have changed homomorphic.go

src/election is for the ballots or votes by users 

src/generator is for generating nodes 

