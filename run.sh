#!/bin/sh

BIN_DIR=bin

CHAIN_DATA_PATH=/home/kairat/diskC/goerli/erigon/chaindata/
BLOCK_INDEX=0
GRAPHVIZ=false
LOOP=false

for i in "$@"; do
    case $i in 
        -p=*|--path=*)
        CHAIN_DATA_PATH="${i#*=}"
        shift
        ;;
        -b=*|--block=*)
        BLOCK_INDEX="${i#*=}"
        shift
        ;;
        -g=*|--graphiz=*)
        GRAPHVIZ="${i#*=}"
        shift
        ;;
        -l=*|--loop=*)
        LOOP="${i#*=}"
        shift
        ;;
        *)
        ;;
    esac
done


mkdir -p $BIN_DIR

go build -o $BIN_DIR/main ./... 

./$BIN_DIR/main -chaindata=$CHAIN_DATA_PATH -block=$BLOCK_INDEX -graphviz=$GRAPHVIZ -loop=$LOOP