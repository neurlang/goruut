# Goruut

A tokenizer, text cleaner, and [IPA](https://en.wikipedia.org/wiki/International_Phonetic_Alphabet) phonemizer/dephonemizer/transphonemizer for several human languages.

## Try it online

It is possible to try this software live at [hashtron.cloud](https://hashtron.cloud/) or at [hugging face](https://huggingface.co/spaces/neurlang/pygoruut).

## Features

* Phone set: IPA
* Supported languages: 87
* Processing speed: fast
* Phone tokens: yes
* Syllable tokens: no
* Word tokens: yes
* Punctuation preservation: yes
* Stressed phones: yes
* Tonal phones: yes, 5 tones (˥ ˦ ˧ ˨ ˩)
* Tie: no

## Installation

```
go install github.com/neurlang/goruut/cmd/goruut@latest
```

## Docker Compose installation

Clone the repo and then run in root directory this command:

```
sudo docker compose up -d --force-recreate --build
```

## Supported Languages

* Afrikaans
* Amharic
* Arabic
* Armenian
* Azerbaijani
* Basque
* Belarusian
* Bengali
* Bengali Dhaka
* Bengali Rahr
* Bulgarian
* Burmese
* Catalan
* Cebuano
* Chechen
* Chichewa
* Chinese Mandarin
* Croatian
* Czech
* Danish
* Dutch
* Dzongkha
* English
* English American
* English British
* Esperanto
* Estonian
* Farsi
* Finnish
* French
* Galician
* Georgian
* German
* Greek
* Gujarati
* Hausa
* Hebrew
* Hindi
* Hungarian
* Icelandic
* Indonesian
* Isan
* Italian
* Jamaican
* Japanese
* Javanese
* Kazakh
* Khmer Central
* Korean
* Lao
* Latvian
* Lithuanian
* Luxembourgish
* Macedonian
* Malayalam
* Malay Arab
* Malay Latin
* Maltese
* Marathi
* Mongolian
* Nepali
* Norwegian
* Pashto
* Polish
* Portuguese
* Punjabi
* Romanian
* Russian
* Serbian
* Slovak
* Spanish
* Swahili
* Swedish
* Tagalog
* Tamil
* Telugu
* Thai
* Tibetan
* Turkish
* Ukrainian
* Urdu
* Uyghur
* Vietnamese Central
* Vietnamese Northern
* Vietnamese Southern
* Yoruba
* Zulu

The goal to support all of [voice2json's languages](https://github.com/synesthesiam/voice2json-profiles#supported-languages) has been met.
However, please [add a language](https://github.com/neurlang/goruut/blob/master/dicts/README.md) if you have the necessary data.

## Listening to the generated speech

There are currently 3 target languages (IPA flavors). They are:

* IPA - Copy the output into [ipa-reader.xyz](http://ipa-reader.xyz/) and pick a correct language voice
* Espeak - Copy the output into espeak. For example czech: `espeak -v cs "[[ru:Zovi: ku:n^]]"`
* Antvaset - Copy the output into [antvaset.com](https://www.antvaset.com/ipa-to-speech) and pick a correct language voice

## Dependencies

See go.mod file for an up-to-date list of depended-on projects. Minimum supported version of golang is go 1.19 (project uses type parameters).

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
Unlike phonemizer, goruut looks up words in a pre-built lexicon (pronunciation dictionary) or guesses word pronunciations with a pre-trained
transformer-based grapheme-to-phoneme model.
