# Hyper parameter search guide

Hyper parameter search is a CPU intensive process. For a specific language
that needs to be solved, usually 10 models are trained. Once they are trained
they can be evaluated how quality is the IPA they produce. I evaluate them
on scale 0 to 10 using a LLM.
Do not train a language which is already solved.
Make sure to train different language than someone else is already training, to not waste effort.

## Preparation

Create some folder. In this folder, clone 3 repositories:

```bash
mkdir hyper_parameter_search
cd hyper_parameter_search
git clone https://github.com/neurlang/goruut
git clone https://github.com/neurlang/dataset
git clone https://github.com/neurlang/classifier
```

Now you need the utterance level phonemizer trainer:

```bash
cd classifier/cmd/train_phonemizer_ulevel/
go build
cd ../../../
```

## Hyper parameter search

Go to the directory `goruut/cmd/analysis2` and run `./hyper_parameter_search.sh <language_folder_name>`:

```bash
cd goruut/cmd/analysis2/
./hyper_parameter_search.sh vietnamese/central
```

## Hyper parameter Evaluation

Go to the directory `goruut/cmd/analysis2` and run `./hyper_parameter_evaluate.sh <language_folder_name>`:

```bash
cd goruut/cmd/analysis2/
./hyper_parameter_search.sh vietnamese/central
```



