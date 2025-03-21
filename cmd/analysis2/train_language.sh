#!/bin/bash

# Initialize a resume flag
resume_flag="-resume"
for arg in "$@"; do
  if [ "$arg" == "-overwrite" ] || [ "$arg" == "--overwrite" ]; then
    resume_flag=""
    break
  fi
done

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
--maxdepth 9999 $resume_flag \
--cleantsv ../../dicts/$1/clean.tsv \
--dstmodel ../../dicts/$1/weights1.json.zlib $2 $3 $4 $5 $6 $7 $8 $9 & # > /dev/null 2>&1 &
PID1=$!

# Start the second process in the background
../backtest/backtest -testing  -langname $1 &
PID2=$!

# Wait for both processes to finish
wait $PID1
cleanup
