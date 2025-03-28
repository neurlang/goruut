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

trap cleanup SIGINT

#train

../../../classifier/cmd/train_phonemizer_multi/train_phonemizer_multi \
--maxdepth 9999 $resume_flag \
--langdir "../../dicts/$lang_name" \
--dstmodel "../../dicts/$lang_name/weights5.json.zlib" \
"${filtered_args[@]}" &  # Pass filtered arguments here
PID1=$!

# Wait for both processes to finish
wait $PID1
cleanup
