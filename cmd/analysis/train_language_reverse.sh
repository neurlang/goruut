#!/bin/bash

#train

../../../classifier/cmd/train_phonemizer/train_phonemizer --cleantsv ../../dicts/$1/clean_reverse.tsv --dstmodel ../../dicts/$1/weights1_reverse.json.lzw $2 $3 $4 $5 $6
