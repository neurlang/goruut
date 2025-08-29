#!/bin/bash
 
# Function to handle SIGINT (Ctrl+C)
cleanup() {
    echo "Caught SIGINT, terminating proces..."
    kill -SIGTERM $PID1 2>/dev/null
}
 
trap cleanup SIGINT
 
lang_name() {
    local dir="$1"
 
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
    local result=""
    IFS='/' read -ra parts <<< "$dir"
    for part in "${parts[@]}"; do
        result+=$(tr '[:lower:]' '[:upper:]' <<< "${part:0:1}")${part:1}
    done
 
    echo "$result"
}
 
lang="$(lang_name $1)"
 
for i in $(seq -10 10); do
 
if [ ! -f ../../dicts/$1/weights6.$i.json.zlib ]; then
    continue
fi
 
rm ../../dicts/$1/weights6.json.zlib
rm ../../dicts/$1/language.json
rm ../../dicts/$1/missing.all.zlib
 
rm ../../dicts/$1/weights4.json.zlib
 
cp ../../dicts/$1/weights6.$i.json.zlib ../../dicts/$1/weights6.json.zlib
cp ../../dicts/$1/language.$i.json ../../dicts/$1/language.json
 
../build.sh
 
 
 
../goruut/goruut --configfile ../../configs/config.json 2> /dev/null &
PID1=$!
 
sleep 1
 
pushd ../../dicts_scripts/
echo -n "$i: ";
./phonemize.sh --lang $lang
popd
 
 
 
cleanup
 
done
