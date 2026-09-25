#!/usr/bin/env bash

lang_name() {
    dir="$1"
    case "$dir" in
        malay/arab) echo "MalayArab"; return;;
        malay/latin) echo "MalayLatin"; return;;
        bengali/dhaka) echo "BengaliDhaka"; return;;
        bengali/rahr) echo "BengaliRahr"; return;;
        bishnupriyamanipuri) echo "BishnupriyaManipuri"; return;;
        chinese/mandarin) echo "ChineseMandarin"; return;;
        english/american) echo "EnglishAmerican"; return;;
        english/british) echo "EnglishBritish"; return;;
        haitiancreole) echo "HaitianCreole"; return;;
        hebrew2) echo "Hebrew2"; return;;
        hebrew3) echo "Hebrew3"; return;;
        khmer/central) echo "KhmerCentral"; return;;
        linguafrancanova) echo "LinguaFrancaNova"; return;;
        lulesaami) echo "LuleSaami"; return;;
        vietnamese/central) echo "VietnameseCentral"; return;;
        vietnamese/northern) echo "VietnameseNorthern"; return;;
        vietnamese/southern) echo "VietnameseSouthern"; return;;
        gaelic/scottish) echo "GaelicScottish"; return;;
        gaelic/irish) echo "GaelicIrish"; return;;
        langbelta) echo "LangBelta"; return;;
        latin/classical) echo "LatinClassical"; return;;
        latin/ecclesiastical) echo "LatinEcclesiastical"; return;;
        nahuatl/classical) echo "NahuatlClassical"; return;;
        nahuatl/central) echo "NahuatlCentral"; return;;
        nahuatl/mecayapan) echo "NahuatlMecayapan"; return;;
        nahuatl/tetelcingo) echo "NahuatlTetelcingo"; return;;
        welsh/north) echo "WelshNorth"; return;;
        welsh/south) echo "WelshSouth"; return;;
        minnan/hokkien) echo "MinnanHokkien"; return;;
        minnan/taiwanese) echo "MinnanTaiwanese"; return;;
        minnan/hokkien2) echo "MinnanHokkien2"; return;;
        minnan/taiwanese2) echo "MinnanTaiwanese2"; return;;
        shantaiyai) echo "ShanTaiYai"; return;;
        mirandese/central) echo "MirandeseCentral"; return;;
    esac

    # Default case: split on '/' and capitalize each part
    result=""
    IFS='/' read -ra parts <<< "$dir"
    for part in "${parts[@]}"; do
        # macOS-compatible capitalization
        first_char=$(echo "${part:0:1}" | tr '[:lower:]' '[:upper:]')
        rest_chars="${part:1}"
        result+="${first_char}${rest_chars}"
    done
    echo "$result"
}

lang="$(lang_name "$1")"

touch "../../dicts/$1/language_reverse.json"

../build.sh

# Use brace expansion instead of seq (works in bash on Windows)
for i in {0..10}; do
    if [ ! -f "../../dicts/$1/weights8_reverse.${i}.bin.zlib" ]; then
        continue
    fi

    rm -f "../../dicts/$1/weights8_reverse.bin.zlib" \
          "../../dicts/$1/language_reverse.json" \
          "../../dicts/$1/missing.all.zlib"

    cp "../../dicts/$1/weights8_reverse.${i}.bin.zlib" "../../dicts/$1/weights8_reverse.bin.zlib"
    cp "../../dicts/$1/language_reverse.${i}.json" "../../dicts/$1/language_reverse.json"

   ../phondephontest/phondephontest --langname "$1" --corpus "$2" --batchsize 99999999 --dictgetterdir ../../dicts/
done

# cleanup
rm -f "../../dicts/$1/weights8_reverse.bin.zlib" \
          "../../dicts/$1/language_reverse.json" \
          "../../dicts/$1/missing.all.zlib"
