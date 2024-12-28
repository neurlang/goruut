#!/bin/bash

# Initialize a reverse flag
reverse_flag=""
for arg in "$@"; do
  if [[ "$arg" == "--reverse" ]]; then
    reverse_flag="_reverse"
    break
  fi
done

analysis_script="./analysis"

original_json="../../dicts/$1/language$reverse_flag.json"
srcfile="../../dicts/$1/dirty.tsv"
dstfile="../../dicts/$1/clean$reverse_flag.tsv"
$analysis_script --target 9999999999999 --lang "$original_json" --srcfile "$srcfile" --dstfile "$dstfile" -loss -nospaced -noipadash $2 $3 $4 $5 $6 $7 $8 $9

