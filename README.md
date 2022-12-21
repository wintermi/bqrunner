# BigQuery Query Runner
[![Go Workflow Status](https://github.com/wintermi/bqrunner/workflows/Go/badge.svg)](https://github.com/wintermi/bqrunner/actions/workflows/go.yml)&nbsp;[![Go Report Card](https://goreportcard.com/badge/github.com/wintermi/bqrunner)](https://goreportcard.com/report/github.com/wintermi/bqrunner)&nbsp;[![license](https://img.shields.io/github/license/wintermi/bqrunner.svg)](https://github.com/wintermi/bqrunner/blob/main/LICENSE)&nbsp;[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/wintermi/bqrunner?include_prereleases)](https://github.com/wintermi/bqrunner/releases)


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
