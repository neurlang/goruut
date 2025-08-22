#!/bin/bash

# Capture language name from first argument
lang_name="$1"
shift  # Remove $1 from arguments list, leaving only parameters for training

# Initialize resume flag and filtered arguments
resume_flag="-resume"
filtered_args=()

# Process remaining arguments ($2 and beyond from original command line)
for arg in "$@"; do
  if [[ "$arg" == "-overwrite" || "$arg" == "--overwrite" ]]; then
    resume_flag=""  # Disable resume if overwrite found
  else
    filtered_args+=("$arg")  # Keep all other arguments
  fi
done

# Function to handle SIGINT (Ctrl+C)
cleanup() {
    echo "Caught SIGINT, terminating processes..."
    kill -SIGTERM $PID1 $PID2 2>/dev/null
    exit 1
}

# Trap the SIGINT signal and call the cleanup function
trap cleanup SIGINT

# Start training process with filtered arguments
../../../classifier/cmd/train_phonemizer_ulevel/train_phonemizer_ulevel \
    --maxdepth 9999 $resume_flag \
    --langdir "../../dicts/$lang_name" \
    --dstmodel "../../dicts/$lang_name/weights6.json.zlib" \
    "${filtered_args[@]}" &  # Pass filtered arguments here
PID1=$!

# Start backtest process
#../backtest/backtest -testing -langname "$lang_name" &
#PID2=$!

# Wait for processes and clean up
wait $PID1
cleanup
