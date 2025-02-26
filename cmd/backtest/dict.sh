#!/bin/bash

./backtest --reverse --dumpwrong --langname $1
./backtest --dumpwrong --langname $1
sort ../../dicts/$1/missing.all.tsv | uniq > /tmp/temp.txt && mv /tmp/temp.txt ../../dicts/$1/missing.all.tsv
./backtest --dumpcompress --langname $1
rm ../../dicts/$1/missing.all.tsv
