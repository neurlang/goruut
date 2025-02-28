#!/bin/bash

./learn.sh $1
cat ../../dicts/$1/learn_reverse.tsv ../../dicts/$1/learn.tsv | sort | uniq > ../../dicts/$1/missing.all.tsv
./backtest --dumpcompress --langname $1
rm ../../dicts/$1/missing.all.tsv
