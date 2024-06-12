# Goruut

A tokenizer, text cleaner, and [IPA](https://en.wikipedia.org/wiki/International_Phonetic_Alphabet) phonemizer for several human languages.

## Installation

```
go install github.com/neurlang/goruut/cmd/goruut@latest
```

## Supported Languages

* Czech
* Slovak

The goal is to support all of [voice2json's languages](https://github.com/synesthesiam/voice2json-profiles#supported-languages)

## Dependencies

See go.mod file for an up-to-date list of depended-on projects. Minimum supported version of golang is go 1.18 (project uses type parameters).

## Numbers, Dates, and More

Unsupported. Please write them using words.

## Command-Line Usage

To start, launch the server using the example config (in configs dir):
```
./goruut -configfile configs/config.json
```
This will launch the server at a specific http port. You should see the port which you specified in the config file:
```
INFO[0000] Binding port: 18080
```
Then you can run queries:

`POST http://127.0.0.1:18080/tts/phonemize/sentence`
```
{
	"Language": "Czech",
	"Sentence": "jsem supr"	
}
```
Output should be:
```
{
	"Words": [
		{
			"Linguistic": "jsem",
			"Phonetic": "jsɛm"
		},
		{
			"Linguistic": "supr",
			"Phonetic": "supr"
		}
	]
}
```
## Intended Audience

goruut is useful for transforming raw text into phonetic pronunciations, similar to [phonemizer](https://github.com/bootphon/phonemizer).
Unlike phonemizer, gruut looks up words in a pre-built lexicon (pronunciation dictionary) or guesses word pronunciations with a pre-trained
grapheme-to-phoneme model.



