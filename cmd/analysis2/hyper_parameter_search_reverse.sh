#!/bin/bash

../build.sh

for n in $(shuf -i 0-10); do
if [[  "$2" == "-resume" ]]; then
if [ -f "../../dicts/$1/weights8_reverse.$n.bin.zlib" ]; then continue; fi
fi
/bin/cp ../../../dataset/$1/l* ../../../goruut/dicts/$1
./study_language_reverse.sh $1 --rowlossimportance $n
./clean_language_reverse.sh $1 --rowlossimportance $n
./train_language_reverse.sh $1 --overwrite
cp ../../dicts/$1/language_reverse.json ../../dicts/$1/language_reverse.$n.json
cp ../../dicts/$1/weights8_reverse.bin.zlib ../../dicts/$1/weights8_reverse.$n.bin.zlib
done
