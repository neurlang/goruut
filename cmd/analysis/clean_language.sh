#!/bin/bash

# Initialize a reverse flag
reverse_flag=""
for arg in "$@"; do
  if [[ "$arg" == "--reverse" ]]; then
    reverse_flag="_reverse"
    break
  fi
done

analysis_script="../analysis2/analysis2"
dicttomap_script="../dicttomap/dicttomap"

original_json="../../dicts/$1/language$reverse_flag.json"
srcfile="../../dicts/$1/dirty.tsv"
dstfile="../../dicts/$1/clean$reverse_flag.tsv"
$analysis_script --lang "$original_json" --srcfile "$srcfile" --dstfile "$dstfile" $2 $3 $4 $5 $6 $7 $8 $9
$dicttomap_script --srcfile "$dstfile" --lang "$original_json" --writeback
