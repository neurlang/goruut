#!/bin/bash

# Paths to the files
mutations_file="../../dicts/$1/$2.language.json"
original_json="../../dicts/$1/language.json"
mutated_json="/tmp/language_mutated.json"
analysis_script="./analysis"
srcfile="../../dicts/$1/dirty.tsv"

# Function to get a random line from a file
function get_random_line() {
    grep -E '^".+],$' "$1" | shuf -n 1
}

# Function to insert a line into the middle of a file
function insert_line_into_file() {
    local line="$1"
    local input_file="$2"
    local output_file="$3 $4 $5 $6 $7 $8"
    local total_lines
    local middle
    total_lines=$(wc -l < "$input_file")
    middle=$((total_lines / 2))
    head -n "$middle" "$input_file" > "$output_file"
    echo "$line" >> "$output_file"
    tail -n +"$((middle + 1))" "$input_file" >> "$output_file"
}

# Run the analysis script
output=$($analysis_script --lang "$original_json" --srcfile "$srcfile" -loss  -nospaced -nostress -noipadash $3 $4 $5 $6 $7 $8)

# Extract the edit distance from the output
prev_edit_distance=$(echo "$output" | grep -oP 'Edit distance is: \K\d+')

echo "Iteration 0: Initial edit distance is $prev_edit_distance"

# Main gradient descent loop
for i in {1..10000000}; do
    # Get a random mutation
    mutation=$(get_random_line "$mutations_file")
    
    # Insert the mutation into the JSON file
    insert_line_into_file "$mutation" "$original_json" "$mutated_json"
    
    # Run the analysis script
    output=$($analysis_script --lang "$mutated_json" --srcfile "$srcfile" -loss  -nospaced -nostress -noipadash $3 $4 $5 $6 $7 $8)
    
    # Extract the edit distance from the output
    edit_distance=$(echo "$output" | grep -oP 'Edit distance is: \K\d+')
    
    # Check if the edit distance has decreased
    if [ "$edit_distance" -lt "$prev_edit_distance" ]; then
        # Keep the mutation
        echo "Iteration $i: Improved edit distance to $edit_distance"
        prev_edit_distance="$edit_distance"
        cp "$mutated_json" "$original_json"
    else
        # Revert to the previous version
        echo "Iteration $i: Edit distance did not improve ($edit_distance >= $prev_edit_distance)"
        cp "$original_json" "$mutated_json"
    fi
done
