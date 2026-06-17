// Package askllmtest implements benchmarking for phonemization using LLM judgement.
package main

import "github.com/neurlang/goruut/lib"
import "github.com/neurlang/goruut/models/requests"
import "github.com/neurlang/goruut/dicts"
import "os"
import "fmt"
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
import "net/http"
import "bytes"
import "strconv"
import "time"

func loop(filename string, top, group int, do func(string)) {
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

	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		line := scanner.Text()
		slice = append(slice, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
	if len(slice) > top {
		slice = slice[:top]
	}

	for n := range slice {
		column := slice[n]

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
	}
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

type dummy struct{}

func (dummy) GetIpaFlavors() map[string]map[string]string {
	return make(map[string]map[string]string)
}
func (dummy) GetPolicyMaxWords() int {
	return 99999999999
}

type chatRequest struct {
	Model           string        `json:"model"`
	Messages        []chatMessage `json:"messages"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

func main() {
	langname := flag.String("langname", "", "directory language name")
	corpus := flag.String("corpus", "", "corpus txt file in language name")
	nostress := flag.Bool("nostress", false, "no stress")
	batchsize := flag.Int("batchsize", 100, "batch size")
	dictgetterdir := flag.String("dictgetterdir", "", "dict getter dir")
	apikey := flag.String("apikey", "", "API key for LLM")
	endpoint := flag.String("endpoint", "http://192.168.1.208:8000/v1", "LLM API base URL")
	model := flag.String("model", "gpt-4o-mini", "LLM model name")
	promptTmpl := flag.String("prompt", "Rate the phonemization quality of the following sentence from 0% to 100% where 0% is completely wrong and 100% is perfectly correct. Return only a number.\nLanguage: {language}\nSentence: {sentence}\nPhonemization: {phonemization}\n", "judge prompt template")
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

	var sum atomic.Uint64
	var count atomic.Uint64

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
		source = strings.Trim(source, " ")
		source = strings.ToLower(source)
		target = strings.ToLower(target)
		words = strings.ToLower(words)

		prompt := strings.ReplaceAll(*promptTmpl, "{sentence}", source)
		prompt = strings.ReplaceAll(prompt, "{phonemization}", target)
		prompt = strings.ReplaceAll(prompt, "{language}", *langname)

		reqBody, _ := json.Marshal(chatRequest{
			Model:           *model,
			ReasoningEffort: "none",
			Messages: []chatMessage{
				{Role: "user", Content: prompt},
			},
		})
		//println("req...", string(reqBody))
		httpReq, err := http.NewRequest("POST", *endpoint+"/chat/completions", bytes.NewReader(reqBody))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if *apikey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+*apikey)
		}

		client := &http.Client{Timeout: 120 * time.Second}
		resp2, err := client.Do(httpReq)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}

		var chatResp chatResponse
		err = json.NewDecoder(resp2.Body).Decode(&chatResp)
		resp2.Body.Close()
		if err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		if len(chatResp.Choices) == 0 {
			fmt.Println("Error: no choices in response")
			return
		}
		//println("resp...", chatResp.Choices[0].Message.Content)
		content := chatResp.Choices[0].Message.Content
		content = strings.TrimSpace(content)
		content = strings.ReplaceAll(content, "%", "")
		content = strings.TrimSpace(content)
		judgement, err := strconv.Atoi(content)
		if err != nil {
			fmt.Println("Error parsing judgement:", err, "raw:", content)
			return
		}
		if judgement < 0 || judgement > 100 {
			fmt.Println("Error: judgement out of range:", judgement)
			return
		}
		sum.Add(uint64(judgement))
		count.Add(1)
	})

	if count.Load() > 0 {
		avg := sum.Load() / count.Load()
		println("[average judgement]", avg, "% for", *langname)
	}
}
