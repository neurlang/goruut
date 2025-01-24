#!/bin/bash

#train

../../../classifier/cmd/train_phonemizer/train_phonemizer --cleantsv ../../dicts/$1/clean.tsv --dstmodel ../../dicts/$1/weights1.json.zlib $2 $3 $4 $5 $6
