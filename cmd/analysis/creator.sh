#!/bin/bash

random=$(shuf -i 1-100000 -n 1)
original_json="../../dicts/$2/language.json"
mutated_json="/tmp/language_mutated.$random.json"
analysis_script="./analysis"
srcfile="../../dicts/$2/dirty.tsv"

# Get initial file size in bytes
initial_size=$(stat -c%s "$original_json")

# Calculate the target size
target_size=$((initial_size * (2 + $1) / (1 + $1)))

# Run the analysis script
output=$($analysis_script --lang "$original_json" --srcfile "$srcfile" -loss  -nospaced -nostress -noipadash $3 $4 $5 $6 $7 $8 --threeway --hits 99999)

# Extract the edit distance from the output
prev_edit_distance=$(echo "$output" | grep -oP 'Edit distance is: \K\d+')
# Extract the hits from the output
init_hits=$(echo "$output" | grep -oP 'Decrease hits to: \K\d+')

echo "Initial edit distance is $prev_edit_distance and initial hits is $init_hits"
cp "$original_json" "$mutated_json"

for ((i = $init_hits; i > 0; i--)); do
    if [[ "$init_hits" -gt 0 ]]; then
        i=$init_hits
    fi


    # Get the current file size
    current_size=$(stat -c%s "$mutated_json")
    # Check if the current size is greater than or equal to the target size
    if (( current_size >= target_size )); then
        echo "File size has increased past the target size. Switching to the other mode."
        break
    fi

    output=$($analysis_script --lang "$mutated_json" --srcfile "$srcfile"  -nospaced -nostress -noipadash $3 $4 $5 $6 $7 $8 --threeway --hits $i --save)
    # Extract the hits from the output
    init_hits=$(echo "$output" | grep -oP 'Decrease hits to: \K\d+')

    # First analysis
    output=$($analysis_script --lang "$mutated_json" --srcfile "$srcfile" -loss  -nospaced -nostress -noipadash $3 $4 $5 $6 $7 $8)
    
    # Extract the edit distance from the output
    edit_distance=$(echo "$output" | grep -oP 'Edit distance is: \K\d+')

    # Check if the edit distance has decreased
    if [ "$edit_distance" -lt "$prev_edit_distance" ]; then
        # Keep the mutation and update the original JSON
        echo "Iteration $i: Improved edit distance to $edit_distance"
        prev_edit_distance="$edit_distance"
        cp "$mutated_json" "$original_json"
    else
        # Revert to the previous version
        echo "Iteration $i: Edit distance did not improve ($edit_distance >= $prev_edit_distance)"
        cp "$original_json" "$mutated_json"
    fi
done
