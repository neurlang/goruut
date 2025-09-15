#!/bin/bash

../build.sh

padspace=""
if [[  "$1" == "cantonese" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "minnan/hokkien" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "minnan/taiwanese" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "minnan/hokkien2" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "minnan/taiwanese2" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "chinese/mandarin" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "japanese" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "tibetan" ]]; then
	padspace="-padspace";
fi
if [[  "$1" == "shantaiyai" ]]; then
	padspace="-padspace";
fi

for n in $(shuf -i 0-10); do
if [[  "$2" == "-resume" ]]; then
if [ -f "../../dicts/$1/weights6.$n.json.zlib" ]; then continue; fi
fi
/bin/cp ../../../dataset/$1/l* ../../../goruut/dicts/$1
./study_language.sh $1 --rowlossimportance $n $padspace
./clean_language.sh $1 --rowlossimportance $n $padspace
./train_language.sh $1 --overwrite
cp ../../dicts/$1/language.json ../../dicts/$1/language.$n.json
cp ../../dicts/$1/weights6.json.zlib ../../dicts/$1/weights6.$n.json.zlib
done
