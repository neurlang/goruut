#!/bin/bash

# Declare an array of strings
string_array=("afrikaans" "amharic" "arabic" "armenian" "azerbaijani" "basque" "belarusian" "bengali" "bulgarian" "burmese" "catalan" "cebuano" "chechen" "chichewa" "croatian" "czech" "danish" "dutch" "dzongkha" "english" "esperanto" "estonian" "farsi" "finnish" "french" "galician" "georgian" "german" "greek" "gujarati" "hausa" "hebrew" "hindi" "hungarian" "icelandic" "indonesian" "isan" "italian" "jamaican" "japanese" "javanese" "kazakh" "khmer/central" "korean" "lao" "latvian" "lithuanian" "luxembourgish" "macedonian" "malay/latin" "malay/arab" "malayalam" "maltese" "marathi" "mongolian" "nepali" "norwegian" "pashto" "polish" "portuguese" "punjabi" "romanian" "russian" "serbian" "slovak" "spanish" "swahili" "swedish" "tagalog" "tamil" "telugu" "thai" "tibetan" "turkish" "ukrainian" "urdu" "uyghur" "vietnamese/northern" "vietnamese/central" "vietnamese/southern" "yoruba" "zulu" "chinese/mandarin" "bengali/dhaka" "bengali/rahr" "english/american" "english/british" "albanian" "aragonese" "assamese" "bashkir" "bishnupriyamanipuri" "bosnian" "cherokee" "chuvash" "gaelic/irish" "gaelic/scottish" "greenlandic" "guarani" "haitiancreole" "hawaiian" "ido" "interlingua" "kannada" "kiche" "konkani" "kurdish" "kyrgyz" "langbelta" "latgalian" "latin/classical" "latin/ecclesiastical" "linguafrancanova" "lojban" "lulesaami" "maori" "nahuatl/central" "nahuatl/classical" "nahuatl/mecayapan" "nahuatl/tetelcingo" "nogai" "oromo" "papiamento" "quechua" "quenya" "setswana" "shantaiyai" "sindarin" "sindhi" "sinhala" "slovenian" "tatar" "turkmen" "uzbek" "welsh/north" "welsh/south" "cantonese" "minnan/taiwanese" "minnan/hokkien" "hebrew2" "hebrew3" "minnan/taiwanese2" "minnan/hokkien2")

LANGFILE='language.json'

# Loop through each string in the array
for LANG in "${string_array[@]}"
do

	echo "{\"Map\":{" > ./$LANG/$LANGFILE
	count=0
	SIZE=0
	while [[ $count -lt 2 && $SIZE -lt 10 ]]
	do
		SIZE=$(($SIZE+1))
		
		# Generate the pattern
		backslash_x=""
		for ((i=0; i<SIZE; i++)); do
		  backslash_x+="\\X"
		done

		# Create the string with the generated \X sequences
		generated_string="\"${backslash_x}\":\[\"[^\"]+\"\],"
		count=$(grep --only-matching -P $generated_string ../goruut/dicts/$LANG/$LANGFILE | wc -l)
		
	done
	echo "$LANG $count $SIZE"
	grep --only-matching -P $generated_string ../goruut/dicts/$LANG/$LANGFILE | grep -P '\xC9\xA1' --invert-match >> ./$LANG/$LANGFILE
	echo '"":[]},' >> ./$LANG/$LANGFILE
	
	
	egrep --only-matching '"(SrcMulti|PrePhonWordSteps|DstMultiSuffix|DstMultiPrefix|DropLast|SplitAfter|SplitBefore|IsDuplex|IsSrcSurround|SrcDuplicate)".+$' ../goruut/dicts/$LANG/$LANGFILE >> ./$LANG/$LANGFILE
	grep --only-matching '^\[.*\][},]$' ../goruut/dicts/$LANG/$LANGFILE >> ./$LANG/$LANGFILE
	#grep DropLast ../goruut/dicts/$LANG/$LANGFILE >> ./$LANG/$LANGFILE

	# Check if the last character is a comma
	# Read the last character (ignoring newlines and whitespace)
	last_char=$(tail -n 1 "./$LANG/$LANGFILE" | tr -d '\n\r' | tail -c 1)
	if [[ "$last_char" == "," ]]; then
	    # Replace the last character with '}'
	    sed -i '$ s/,$/}/' "./$LANG/$LANGFILE"
	    echo "Replaced last comma with '}'."
	else
	    echo "Last character is not a comma but $last_char. No changes made."
	fi
done
