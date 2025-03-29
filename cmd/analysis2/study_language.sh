#!/bin/bash

# Initialize a reverse flag
reverse_flag=""
for arg in "$@"; do
  if [[ "$arg" == "--reverse" ]]; then
    reverse_flag="_reverse"
    break
  fi
done

analysis_script="./analysis2"

original_json="../../dicts/$1/language$reverse_flag.json"
srcfile="../../dicts/$1/lexicon.tsv"
$analysis_script --lang "$original_json" --srcfile "$srcfile" -deleting -threeway -save  $2 $3 $4 $5 $6 $7 $8 $9

