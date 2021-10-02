# fj

Flatten your JSON.

`fj` flattens JSON into a value and its full path per line, so data can be worked on using UNIX tools such as grep/awk/cut/etc.

<pre>
❯ curl -s "https://api.github.com/repos/golang/go/commits?per_page=1" | <b>fj</b> | grep commit.author
json[0].commit.author.name  "Josh Bleecher Snyder"
json[0].commit.author.email "josharian@gmail.com"
json[0].commit.author.date  "2021-06-30T16:44:30Z"
</pre>

fj can un-flatten too, which is useful to get a subset of JSON:

<pre>
❯ curl -s "https://api.github.com/repos/golang/go/commits?per_page=1" | fj | grep commit.author | <b>fj -u</b> | jq
[
  {
    "commit": {
      "author": {
        "name": "Josh Bleecher Snyder",
        "email": "josharian@gmail.com",
        "date": "2021-06-30T16:44:30Z"
      }
    }
  }
]
</pre>

## Installation

1. Download [latest release for Linux, Mac, Windows or FreeBSD](https://github.com/pitr/fj/releases),
2. Put the binary in your `$PATH` (e.g. in `/usr/local/bin`) to make it easy to use:

Alternatively, if you have Go compiler:

```
❯ go get -u github.com/pitr/fj
```

## Usage

Flatten JSON file or stdin:

```
❯ fj file.json
❯ cat file.json | fj
```

Or many JSON files. Use `-s` or `--stream` to treat input as a stream of JSON documents:

```
❯ fj -s file1.json file2.json file3.json
```

Use `grep` to search, `diff <(fj file1.json) <(fj file2.json)` to diff, and other  tools such as `awk/sort/uniq/fzf`.

## FAQ

### Why shouldn't I just use gron?

[gron](https://github.com/tomnomnom/gron) is a very similar tool. However, `fj` is different from it in a few key ways:

- fj does not keep the whole JSON object in memory, which allows it to be 10-20 times faster than gron.

  ```
  ❯ hyperfine --warmup 5 'gron testdata/big.json' 'fj testdata/big.json'
  Benchmark #1: gron testdata/big.json
    Time (mean ± σ):     136.2 ms ±   1.9 ms    [User: 57.1 ms, System: 97.2 ms]
    Range (min … max):   132.9 ms … 140.2 ms    21 runs

  Benchmark #2: fj testdata/big.json
    Time (mean ± σ):       9.1 ms ±   0.9 ms    [User: 4.9 ms, System: 2.4 ms]
    Range (min … max):     7.8 ms …  12.2 ms    203 runs

  Summary
    'fj testdata/big.json' ran
     14.93 ± 1.48 times faster than 'gron testdata/big.json'
  ```

- `fj` does not preserve array structures by padding with null.

  ```
  ❯ echo '[1,2,3,4,5]' | gron | grep 5 | gron -u
  [
     null,
     null,
     null,
     null,
     5
  ]
  ❯ echo '[1,2,3,4,5]' | fj | grep 5 | fj -u
  [5]
  ```

- `fj` does not try to convert JSON into valid JavaScript that can be pasted into JS console. There are no extra semicolons and array/object initializations.

### Why shouldn't I just use jq?
[jq](https://stedolan.github.io/jq/) is a great and powerful tool with its own language. `fj` simply flattens (and un-flattens) JSON, and is expected to integrate with existing tools.
