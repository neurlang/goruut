#!/bin/bash

: > ../../dicts/$1/learn_reverse.tsv
: > ../../dicts/$1/learn.tsv
./backtest --reverse --dumpwrong --langname $1
./backtest --dumpwrong --langname $1
