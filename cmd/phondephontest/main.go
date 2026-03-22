// Package backtest implements benchmarking for phonemization on unlabelled data.
package main

import "github.com/neurlang/levenshtein"
import "github.com/neurlang/goruut/lib"
import "github.com/neurlang/goruut/models/requests"
import "github.com/neurlang/goruut/dicts"
import "os"
import "fmt"
import "github.com/neurlang/classifier/parallel"
import "bufio"
import "flag"
import "strings"
import "sync/atomic"
import "math/rand"
import "github.com/klauspost/compress/zstd"
import "io"
import "encoding/json"
import di "github.com/martinarisk/di/dependency_injection"
import "github.com/neurlang/goruut/repo/interfaces"

func loop(filename string, top, group int, do func(string)) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var rdr = io.ReadCloser(file)

	if strings.HasSuffix(filename, ".zst") || strings.HasSuffix(filename, ".zstd") {
		r, err := zstd.NewReader(file)
		if err != nil {
			fmt.Println("Error decompressing file:", err)
			return
		}
		rdr = r.IOReadCloser()
	}

	var slice []string

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()
		slice = append(slice, line)
	}

	// Check for any scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
	if len(slice) > top {
		slice = slice[:top]
	}

	parallel.ForEach(len(slice), group, func(n int) {
		// Process each column
		column := slice[n]
		// Example: Print the column

		if strings.Contains(filename, ".json") {
			var buf map[string]string
			err := json.Unmarshal([]byte(column), &buf)
			if err == nil {
				do(buf["text"])
			} else {
				fmt.Println("Error parsing json:", err)
			}
		} else {
			do(column)
		}
	})
}


type DictGetter struct {
	getter      dicts.DictGetter
	coolname    string
                langname    string
	dir         string
}

func (d *DictGetter) GetDict(lang, filename string) ([]byte, error) {
	if lang == d.coolname {
		fullpath := d.dir + string(os.PathSeparator) + d.langname + string(os.PathSeparator) + filename
		data, err := os.ReadFile(fullpath)
		return data, err
	}
	return d.getter.GetDict(lang, filename)
}
func (d *DictGetter) IsNewFormat(magic []byte) bool {
	return true
}
func (d *DictGetter) IsOldFormat(magic []byte) bool {
	return false
}


type dummy struct {
}

func (dummy) GetIpaFlavors() map[string]map[string]string {
	return make(map[string]map[string]string)
}
func (dummy) GetPolicyMaxWords() int {
	return 99999999999
}

func main() {
	langname := flag.String("langname", "", "directory language name")
	corpus := flag.String("corpus", "", "corpus txt file in language name")
	nostress := flag.Bool("nostress", false, "no stress")
	batchsize := flag.Int("batchsize", 100, "batch size")
	dictgetterdir := flag.String("dictgetterdir", "", "dict getter dir")
	flag.Parse()

	if corpus == nil || *corpus == "" {
		println("ERROR: Corpus flag is mandatory")
		return
	}
	var dictgetter DictGetter
	var coolname string
	if langname != nil {
		coolname = dicts.LangName(*langname)
		dictgetter.coolname = coolname
		dictgetter.langname = *langname
	}
	p := lib.NewPhonemizer(nil)
	if dictgetterdir != nil && "" != *dictgetterdir {
		dictgetter.dir = *dictgetterdir
		di := di.NewDependencyInjection()
		di.Add((interfaces.DictGetter)(&dictgetter))
		di.Add((interfaces.IpaFlavor)(dummy{}))
		di.Add((interfaces.PolicyMaxWords)(dummy{}))
		p = lib.NewPhonemizer(di)
	}
	var percent, errsum, total, maxsum atomic.Uint64
	loop(*corpus, *batchsize, 1000, func(words string) {
		if nostress != nil && *nostress {
			words = strings.ReplaceAll(words, "'", "")
			words = strings.ReplaceAll(words, "ˈ", "")
			words = strings.ReplaceAll(words, "ˌ", "")
		}
		resp := p.Sentence(requests.PhonemizeSentence{
			Sentence:  words,
			Language:  coolname,
			IsReverse: false,
		})
		var source string
		var target string
		for i := range resp.Words {
			source += resp.Words[i].CleanWord + " "
			target += resp.Words[i].Phonetic + " "
		}
		target = strings.Trim(target, " ")
		if nostress != nil && *nostress {
			target = strings.ReplaceAll(target, "'", "")
			target = strings.ReplaceAll(target, "ˈ", "")
			target = strings.ReplaceAll(target, "ˌ", "")
		}
		resp = p.Sentence(requests.PhonemizeSentence{
			Sentence:  target,
			Language:  coolname,
			IsReverse: true,
		})
		target = ""
		for i := range resp.Words {
			target += resp.Words[i].Phonetic + " "
		}
		target = strings.Trim(target, " ")
		if nostress != nil && *nostress {
			target = strings.ReplaceAll(target, "'", "")
			target = strings.ReplaceAll(target, "ˈ", "")
			target = strings.ReplaceAll(target, "ˌ", "")
		}
		source = strings.Trim(source, " ")
		source = strings.ToLower(source)
		target = strings.ToLower(target)
		words = strings.ToLower(words)

		var dist = *levenshtein.Distance(levenshtein.Matrix[uint64](uint(len([]rune(target))), uint(len([]rune(words))),
			nil, nil,
			levenshtein.OneSlice[rune, uint64]([]rune(target), []rune(words)), nil))
		var maxdist = len(target) + len(words)
		errsum.Add(dist)
		maxsum.Add(uint64(maxdist))
		source_split := strings.Split(source, " ")
		target_split := strings.Split(target, " ")
		var dist2 = *levenshtein.Distance(levenshtein.Matrix[uint64](uint(len(target_split)), uint(len(source_split)),
			nil, nil,
			levenshtein.OneSlice[string, uint64](target_split, source_split), nil))
		for i := 0; i < len(source_split) || i < len(target_split); i++ {
			total.Add(1)
		}
		percent.Add(dist2)
	})
	if total.Load() > 0 {
		success_wer := 100 - (100 * percent.Load() / total.Load())
		println("[success rate WER]", success_wer, "%", percent.Load(), "for", *langname)
	}
	if maxsum.Load() > 0 {
		success_cer := 100 - (100 * errsum.Load() / maxsum.Load())
		println("[success rate CER]", success_cer, "%", errsum.Load(), "for", *langname)
	}
}
