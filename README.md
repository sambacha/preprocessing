# Erigon block transactions analysis

## Goal
The goal is pretty much simple - to find out if transactions in a block can be executed independently. What does it mean - independently? It means that every transaction in block does not change account states on which other transactions in the same block are dependant. Example:
![Dependant transactions](/images/depend_exec_1.jpg)<br />
In this simple example `Transaction 1` changes the state of an account with address `B`, while `Transaction 2` and `Transaction 3` require to know the state of an account to proceed with code execution. Thus `Transactions 2 and 3` depend on `Transaction 1`. See [here for more detailed explanation.](/docs/01_transactions.md) 

## Usage
Using `run.sh` bash script. It requires to change the path to chain database 
```sh
# run.sh
CHAIN_DATA_PATH=/path/to/chaindata
```
```
./run.sh -b=156893 -g=true
```

Script flags:
```
-b|--block=<uint> (default 0) - block number to analize
-l|--loop=<bool> (default false) - perform a loop starting from block number? if flag is true, -g flag is always false
-g|--graphviz=<bool> (default false) - generate visual representation of bytecode?
-p|--path=<string> (default CHAIN_DATA_PATH) - path to chain database
```
Using `make`. It requires to change `DEFAULT_PATH` in `main.go`.
```
make build
```
```
./bin/main -block=1234 -loop
```



