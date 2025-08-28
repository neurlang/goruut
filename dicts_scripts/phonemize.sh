# USEAGE: ./phonemize.sh --lang EnglishAmerican
#!/bin/bash

# Default values
TRANSCRIPT_FILE="transcript.json"
declare -a LANGUAGES  # Array to store allowed languages

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --lang)
            LANGUAGES+=("$2")
            shift 2
            ;;
        *)
            echo "Unknown parameter: $1"
            exit 1
            ;;
    esac
done

# Read each entry from the JSON and process it
jq -r 'to_entries[] | "\(.key) \(.value)"' "$TRANSCRIPT_FILE" | while read -r key sentence; do
    # Skip processing if specific languages were specified and this key doesn't match
    if [[ ${#LANGUAGES[@]} -gt 0 ]]; then
        found=0
        for lang in "${LANGUAGES[@]}"; do
            if [[ "$key" == "$lang" ]]; then
                found=1
                break
            fi
        done
        [[ $found -eq 0 ]] && continue
    fi

    # Send POST request
    response=$(curl -s -X POST http://127.0.0.1:18080/tts/phonemize/sentence \
        -H "Content-Type: application/json" \
        -d "{\"Language\": \"$key\", \"Sentence\": \"$sentence\"}")

    # Extract each word's PrePunct + Phonetic + PostPunct
    phonetics=$(echo "$response" | jq -r '.Words[] | "\(.PrePunct)\(.Phonetic)\(.PostPunct)"' | paste -sd ' ' -)

    # Print the formatted phonetic line
    echo "$phonetics"
done
