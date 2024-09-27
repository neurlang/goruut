#!/bin/bash

analysis_script="./analysis"


# Declare an array of strings
string_array=("czech" "slovak" "romanian" "finnish" "isan" "swahili" "esperanto" "icelandic" "norwegian" "jamaican")

# Loop through each string in the array
for lang in "${string_array[@]}"
do
	original_json="../../dicts/$lang/language.json"
	srcfile="../../dicts/$lang/dirty.tsv"
	dstfile="../../dicts/$lang/clean.tsv"
	$analysis_script --lang "$original_json" --srcfile "$srcfile" --dstfile "$dstfile" -loss -randomize 0 -nospaced -nostress -noipadash
done

# Declare an array of strings
string_array=("japanese")

# Loop through each string in the array
for lang in "${string_array[@]}"
do
	original_json="../../dicts/$lang/language.json"
	srcfile="../../dicts/$lang/dirty.tsv"
	dstfile="../../dicts/$lang/clean.tsv"
	$analysis_script --lang "$original_json" --srcfile "$srcfile" --dstfile "$dstfile" -loss -randomize 0 -nospaced -nostress -noipadash -padspace
done


