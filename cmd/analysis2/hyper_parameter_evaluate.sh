#!/usr/bin/env bash

# Function to handle SIGINT (Ctrl+C)
cleanup() {
    echo "Caught SIGINT, terminating process..."
    [ -n "$PID1" ] && kill -SIGTERM "$PID1" 2>/dev/null
    exit 1
}
trap cleanup INT

stop_goruut() {
    [ -n "$PID1" ] && kill -SIGTERM "$PID1" 2>/dev/null
    PID1=""
}

lang_name() {
    dir="$1"
    case "$dir" in
        malay/arab) echo "MalayArab"; return;;
        malay/latin) echo "MalayLatin"; return;;
        bengali/dhaka) echo "BengaliDhaka"; return;;
        bengali/rahr) echo "BengaliRahr"; return;;
        chinese/mandarin) echo "ChineseMandarin"; return;;
        english/american) echo "EnglishAmerican"; return;;
        english/british) echo "EnglishBritish"; return;;
        hebrew2) echo "Hebrew2"; return;;
        hebrew3) echo "Hebrew3"; return;;
        khmer/central) echo "KhmerCentral"; return;;
        vietnamese/central) echo "VietnameseCentral"; return;;
        vietnamese/northern) echo "VietnameseNorthern"; return;;
        vietnamese/southern) echo "VietnameseSouthern"; return;;
        gaelic/scottish) echo "GaelicScottish"; return;;
        gaelic/irish) echo "GaelicIrish"; return;;
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

# Use brace expansion instead of seq (works in bash on Windows)
for i in {0..10}; do
    if [ ! -f "../../dicts/$1/weights6.${i}.json.zlib" ]; then
        continue
    fi

    rm -f "../../dicts/$1/weights6.json.zlib" \
          "../../dicts/$1/language.json" \
          "../../dicts/$1/missing.all.zlib" \
          "../../dicts/$1/weights4.json.zlib"

    cp "../../dicts/$1/weights6.${i}.json.zlib" "../../dicts/$1/weights6.json.zlib"
    cp "../../dicts/$1/language.${i}.json" "../../dicts/$1/language.json"

    ../build.sh

    ../goruut/goruut --configfile ../../configs/config.json 2>/dev/null &
    PID1=$!

    sleep 1

    (
        cd ../../dicts_scripts/ || exit 1
        printf "%s: " "${i}"
        ./phonemize.sh --lang "$lang"
    )

    sleep 1

    # free the port before next iteration
    stop_goruut
done
