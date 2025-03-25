# Example usage

```bash
./phondephontest -langname hebrew2 -corpus ../../../high-quality-multilingual-sentences/he.jsonl.zst --batchsize 999999999
```

## Cloning the high-quality-multilingual-sentences dataset

install `git-lfs`, then:

```bash
git clone https://huggingface.co/datasets/agentlans/high-quality-multilingual-sentences
git lfs fetch
git lfs clone
git lfs pull
```

## Example result

```
[success rate WER] 97 % 561258 for hebrew2
[success rate CER] 18 % 3031252 for hebrew2
```
