#!/bin/bash

# Paths to the files
random=$(shuf -i 1-100000 -n 1)
original_json="../../dicts/$2/language.json"
mutated_json="/tmp/language_mutated.$random.json"
analysis_script="./analysis"
srcfile="../../dicts/$2/dirty.tsv"

# Function to insert a line into the middle of a file
function delete_line_from_file() {
	local input_file="$1"
	local output_file="$2"

	# Find lines matching the regex and shuffle them to get a random one
	LINE_TO_DELETE=$(grep -En '^".+],$' "$input_file" | shuf -n 1 | cut -d: -f1)

	cp "$input_file" "$output_file"

	# Check if a line was found and delete it
	if [ -n "$LINE_TO_DELETE" ]; then
	    sed -i "${LINE_TO_DELETE}d" "$output_file"
	else
	    echo "No matching line found."
	fi
}

# Run the analysis script
output=$($analysis_script --lang "$original_json" --srcfile "$srcfile" -loss  -nospaced -noipadash $3 $4 $5 $6 $7 $8 $9)

# Extract the edit distance from the output
prev_edit_distance=$(echo "$output" | grep -oP 'Edit distance is: \K\d+')

echo "Iteration 0: Initial edit distance is $prev_edit_distance"

j=0

# Main gradient descent loop
for i in {1..10000000}; do
    if (( j >= 10 * ($1 + 3) )); then
	echo "No improvement possible. Switching to the other mode."
	break
    fi
    # Insert the mutation into the JSON file
    delete_line_from_file "$original_json" "$mutated_json"
    
    # Run the analysis script
    output=$($analysis_script --lang "$mutated_json" --srcfile "$srcfile" -loss  -nospaced -noipadash $3 $4 $5 $6 $7 $8 $9)
    
    # Extract the edit distance from the output
    edit_distance=$(echo "$output" | grep -oP 'Edit distance is: \K\d+')
    
    # Check if the edit distance has decreased
    if [ "$edit_distance" -lt "$prev_edit_distance" ]; then
        # Keep the mutation
        echo "Iteration $i: Round $j: Improved edit distance to $edit_distance"
        prev_edit_distance="$edit_distance"
        cp "$mutated_json" "$original_json"
        j=0
    else
        # Revert to the previous version
        echo "Iteration $i: Round $j: Edit distance did not improve ($edit_distance >= $prev_edit_distance)"
        cp "$original_json" "$mutated_json"
    fi

    # Insert the mutation into the JSON file
    cp "$original_json" "$mutated_json"
    
    # Run the analysis script
    output=$($analysis_script --lang "$mutated_json" --srcfile "$srcfile" -loss  -nospaced -noipadash $3 $4 $5 $6 $7 $8 $9 -deleteval -save)
    
    # Extract the edit distance from the output
    edit_distance=$(echo "$output" | grep -oP 'Edit distance is: \K\d+')
    
    # Check if the edit distance has decreased
    if [ "$edit_distance" -lt "$prev_edit_distance" ]; then
        # Keep the mutation
        echo "Iteration $i: Round $j: Improved edit distance to $edit_distance"
        prev_edit_distance="$edit_distance"
        cp "$mutated_json" "$original_json"
        j=0
    else
        # Revert to the previous version
        echo "Iteration $i: Round $j: Edit distance did not improve ($edit_distance >= $prev_edit_distance)"
        cp "$original_json" "$mutated_json"
    fi
    
    j=$((j+1))
done
