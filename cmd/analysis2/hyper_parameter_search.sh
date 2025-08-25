#!/bin/bash

for n in $(shuf -i 0-10); do
/bin/cp ../../dataset/$1/l* ../../../goruut/dicts/$1
./study_language.sh $1 --rowlossimportance $n
./clean_language.sh $1 --rowlossimportance $n
./train_language.sh $1 --overwrite
cp ../../dicts/$1/language.json ../../dicts/$1/language.$n.json
cp ../../dicts/$1/weights6.json.zlib ../../dicts/$1/weights6.$n.json.zlib
done
