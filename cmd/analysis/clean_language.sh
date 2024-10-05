#!/bin/bash

analysis_script="./analysis"

original_json="../../dicts/$1/language.json"
srcfile="../../dicts/$1/dirty.tsv"
dstfile="../../dicts/$1/clean.tsv"
$analysis_script --target 9999999999999 --lang "$original_json" --srcfile "$srcfile" --dstfile "$dstfile" -loss -nospaced -noipadash $2 $3 $4 $5 $6 $7 $8 $9

