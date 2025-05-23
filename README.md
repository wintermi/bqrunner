# BigQuery Query Runner

[![Workflows](https://github.com/wintermi/bqrunner/workflows/Go%20-%20Build/badge.svg)](https://github.com/wintermi/bqrunner/actions)
[![Go Report](https://goreportcard.com/badge/github.com/wintermi/bqrunner)](https://goreportcard.com/report/github.com/wintermi/bqrunner)
[![License](https://img.shields.io/github/license/wintermi/bqrunner.svg)](https://github.com/wintermi/bqrunner/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/wintermi/bqrunner?include_prereleases)](https://github.com/wintermi/bqrunner/releases)


## Description

A command line application designed to provide a simple method to execute one or more SQL queries against a given dataset in BigQuery.  A detailed log is output to the console providing you with the available execution statistics.

```
USAGE:
    bqrunner -p PROJECT_ID -d DATASET -i INPUT_PATH -o OUTPUT_PATH

ARGS:
  -c	Disable Query Cache
  -d string
    	BigQuery Dataset  (Required)
  -dr
    	Dry Run
  -f string
    	Field Delimter (default ",")
  -i string
    	Input SQL Path  (Required)
  -l string
    	BigQuery Data Processing Location
  -o string
    	Output Results Path  (Required)
  -p string
    	Google Cloud Project ID  (Required)
  -s	Shuffle Queries
  -v	Output Verbose Detail
```


## License

**bqrunner** is released under the [Apache License 2.0](https://github.com/wintermi/bqrunner/blob/main/LICENSE) unless explicitly mentioned in the file header.
