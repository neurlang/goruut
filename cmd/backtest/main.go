package main

import "github.com/neurlang/levenshtein"
import "github.com/neurlang/goruut/lib"
import "github.com/neurlang/goruut/dicts"
import "github.com/neurlang/goruut/models/requests"
import di "github.com/martinarisk/di/dependency_injection"
import "github.com/neurlang/goruut/repo/interfaces"
import "os"
import "fmt"
import "github.com/neurlang/classifier/parallel"
import "bufio"
import "flag"
import "strings"
import "sync/atomic"
import "math/rand"
import "time"

func loop(filename string, top, group int, do func(string, string)) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var slice [][2]string

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")

		// Check if we have exactly two or three columns
		if len(columns) != 2 && len(columns) != 3 {
			fmt.Println("Line does not have exactly two or three columns:", line)
			continue
		}

		// Process each column
		column1 := columns[0]
		column2 := columns[1]

		slice = append(slice, [2]string{column1, column2})
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
		column1 := slice[n][0]
		column2 := slice[n][1]
		// Example: Print the columns
		do(column1, column2)
	})

}

type DictGetter struct {
	getter      dicts.DictGetter
	coolname    string
	modelfile   string
	currentfile []byte
	bestfile    []byte
	bestsuccess uint64
}

func (d *DictGetter) GetDict(lang, filename string) ([]byte, error) {
	if lang == d.coolname && strings.HasSuffix(d.modelfile, filename) {
		data, err := os.ReadFile(d.modelfile)
		if err == nil {
			d.currentfile = data
		}
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
func (d *DictGetter) Write() {
	os.WriteFile(d.modelfile+".best", d.bestfile, 0777)
}

func watchFile(filePath string) error {
    initialStat, err := os.Stat(filePath)
    if err != nil {
    	println(err.Error())
        return err
    }

    for {
        stat, err := os.Stat(filePath)
        if err != nil {
    	    println(err.Error())
            return err
        }

        if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
        	println("changed")
            break
        }

        time.Sleep(100 * time.Millisecond)
    }

    return nil
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
	isreverse := flag.Bool("reverse", false, "is reverse")
	nostress := flag.Bool("nostress", false, "no stress")
	testing := flag.Bool("testing", false, "keep backtesting and overwriting the model with the best one")
	resume := flag.Bool("resume", false, "test old model initially")
	flag.Parse()

	var dictgetter DictGetter
	var coolname string
	var srcfile string
	var modelfile string
again:
	if langname != nil {
		coolname = dicts.LangName(*langname)
		dictgetter.coolname = coolname
		srcfile = "../../dicts/" + *langname + "/dirty.tsv"
		if testing != nil && *testing {
			if isreverse != nil && *isreverse {
				modelfile = "../../dicts/" + *langname + "/weights2_reverse.json.zlib"
			} else {
				modelfile = "../../dicts/" + *langname + "/weights2.json.zlib"
			}
			dictgetter.modelfile = modelfile
			if resume != nil && *resume {
				dictgetter.modelfile += ".best"
			}
		}
	}
	var batch = 10000
	p := lib.NewPhonemizer(nil)
	if testing != nil && *testing {
		batch = 1000
		di := di.NewDependencyInjection()
		di.Add((interfaces.DictGetter)(&dictgetter))
		di.Add((interfaces.IpaFlavor)(dummy{}))
		di.Add((interfaces.PolicyMaxWords)(dummy{}))
		p = lib.NewPhonemizer(di)
	}

	var percent, errsum, total atomic.Uint64
	loop(srcfile, batch, 1000, func(word1, word2 string) {
		total.Add(1)
		if nostress != nil && *nostress {
			word2 = strings.ReplaceAll(word2, "'", "")
			word2 = strings.ReplaceAll(word2, "ˈ", "")
			word2 = strings.ReplaceAll(word2, "ˌ", "")
		}

		if isreverse != nil && *isreverse {
			word1, word2 = word2, word1
		}
		resp := p.Sentence(requests.PhonemizeSentence{
			Sentence:  word1,
			Language:  coolname,
			IsReverse: isreverse != nil && *isreverse,
		})

		var target string
		for i := range resp.Words {
			target += resp.Words[i].Phonetic + " "
		}
		target = strings.Trim(target, " ")
		if nostress != nil && *nostress {
			target = strings.ReplaceAll(target, "'", "")
			target = strings.ReplaceAll(target, "ˈ", "")
			target = strings.ReplaceAll(target, "ˌ", "")
		}
		target = strings.ToLower(target)
		word2 = strings.ToLower(word2)

		var mat = levenshtein.Matrix[uint64](uint(len([]rune(target))), uint(len([]rune(word2))),
			nil, nil,
			levenshtein.OneSlice[rune, uint64]([]rune(target), []rune(word2)), nil)
		var dist = *levenshtein.Distance(mat)
		errsum.Add(dist)
		if target == word2 {
			percent.Add(1)
		}

		//success := 100 * percent.Load() / total.Load()
		//println("[success rate]", success, "%", "with", errsum.Load(), "errors", percent.Load(), "successes", "for", *langname)
	})
	success := 100 * percent.Load() / total.Load()
	println("[success rate]", success, "%", "with", errsum.Load(), "errors", percent.Load(), "successes", "for", *langname)

	if testing != nil && *testing {

		if success > dictgetter.bestsuccess {
			dictgetter.bestfile = dictgetter.currentfile
			dictgetter.bestsuccess = success
			dictgetter.Write()
		}
		watchFile(dictgetter.modelfile)
		time.Sleep(time.Second)
		
		dictgetter.modelfile = modelfile
		goto again
	}
}
