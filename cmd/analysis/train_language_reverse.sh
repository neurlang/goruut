#!/bin/bash

# Function to handle the SIGINT signal (Ctrl+C)
cleanup() {
    echo "Caught SIGINT, killing both processes..."
    kill -SIGTERM $PID1 $PID2
    exit 1
}

# Trap the SIGINT signal and call the cleanup function
trap cleanup SIGINT

#train

../../../classifier/cmd/train_phonemizer/train_phonemizer \
--cleantsv ../../dicts/$1/clean_reverse.tsv \
--dstmodel ../../dicts/$1/weights1_reverse.json.zlib $2 $3 $4 $5 $6 $7 $8 $9 & # > /dev/null 2>&1 &
PID1=$!

# Start the second process in the background
../backtest/backtest -reverse -testing  -langname $1 &
PID2=$!

# Wait for both processes to finish
wait $PID1
cleanup
