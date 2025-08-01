#!/bin/bash

EXTRAFILE=''
if test -f ../../dicts/$1/abbr.tsv; then
  sed  -e 's/"/""/g' -e 's/\[/"\[/g' -e 's/\]/\]"/g' "../../dicts/$1/abbr.tsv" > "../../dicts/$1/abbr.dq.tsv"
  EXTRAFILE="../../dicts/$1/abbr.dq.tsv"
fi

./learn.sh $1



cat ../../dicts/$1/learn_reverse.tsv ../../dicts/$1/learn.tsv $EXTRAFILE | sort | uniq > ../../dicts/$1/missing.all.tsv
./backtest --dumpcompress --langname $1
rm ../../dicts/$1/missing.all.tsv

if test -f  "$EXTRAFILE"; then
  rm $EXTRAFILE
fi
