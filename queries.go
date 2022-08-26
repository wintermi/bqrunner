// Copyright 2022, Matthew Winter
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type QueryDetails struct {
	SQL                  string
	InputFile            string
	OutputFile           string
	Number               int
	Error                error
	QueryStartTime       time.Time
	QueryEndTime         time.Time
	FirstRowReturnedTime time.Time
	AllRowsReturnedTime  time.Time
	TotalRowsReturned    int64
}

type Queries struct {
	Query          []QueryDetails
	ExecutionOrder []int
}

//---------------------------------------------------------------------------------------

// Walk the provided Input Path and load all SQL Query Files
func (queries *Queries) LoadQueries(inputPath string, outputPath string) error {

	// Calculate the Absolute Output Path and add a results sub directory
	outputPath, err := filepath.Abs(filepath.Join(outputPath, time.Now().Format("2006-01-02.150405")))
	if err != nil {
		return fmt.Errorf("[LoadQueries] Failed To Get Absolute Output Path: %w", err)
	}

	// Execute a Glob to return all files matching the provided pattern
	matches, err := filepath.Glob(inputPath)
	if err != nil {
		return fmt.Errorf("[LoadQueries] Glob Failed: %w", err)
	}

	// Load all matching files returned from the Glob
	queryCount := 1
	for _, filename := range matches {
		fileInfo, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("[LoadQueries] Failed To Get File Info: %w", err)
		}

		// Skip Directories
		if fileInfo.IsDir() {
			continue
		}

		inputFile, _ := filepath.Abs(filename)
		outputFile, _ := filepath.Abs(filepath.Join(outputPath, fmt.Sprintf("results-query-%06d.output", queryCount)))

		buf, err := ioutil.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("[LoadQueries] Read Input File Failed: %w", err)
		}

		sql := QueryDetails{
			SQL:        string(buf),
			InputFile:  inputFile,
			OutputFile: outputFile,
			Number:     queryCount,
		}

		logger.Debug().Int("Query Number", sql.Number).Msg("Query Details")
		logger.Debug().Str("Input File", sql.InputFile).Msg(indent)
		logger.Debug().Str("Output File", sql.OutputFile).Msg(indent)
		logger.Debug().Str("SQL", sql.SQL).Msg(indent)

		queries.Query = append(queries.Query, sql)
		queries.ExecutionOrder = append(queries.ExecutionOrder, queryCount-1)
		queryCount++
	}

	return nil
}

//---------------------------------------------------------------------------------------

// Shuffle the Query Execution Order
func (queries *Queries) ShuffleExecutionOrder(shuffle bool) {
	if shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(queries.ExecutionOrder), func(i, j int) {
			queries.ExecutionOrder[i], queries.ExecutionOrder[j] = queries.ExecutionOrder[j], queries.ExecutionOrder[i]
		})

		logger.Info().Msg("Query Execution Order Shuffle Complete")
	}
}

//---------------------------------------------------------------------------------------

// Execute the Queries in BigQuery
func (queries *Queries) ExecuteQueries(project string, dataset string, location string, disableQueryCache bool, dryRun bool, delimiter string) error {

	// Establish a BigQuery Client Connection
	logger.Info().Msg("Establishing BigQuery Client Connection")
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("Failed Establishing BigQuery Client Connection: %w", err)
	}
	defer client.Close()

	// BigQuery Client Configuration
	client.Location = location

	// Execut SQL Queries
	errorCount := 0
	for _, index := range queries.ExecutionOrder {
		qd := queries.Query[index]

		if dryRun {
			qd.ExecuteDryRun(ctx, client, project, dataset, location, disableQueryCache)
			qd.LogExecuteDryRun()
		} else {
			qd.ExecuteQuery(ctx, client, project, dataset, location, disableQueryCache, delimiter)
			qd.LogExecuteQuery()
		}

		// Count the Errors
		if qd.Error != nil {
			errorCount++
		}
	}

	// Raise an Error if one of the queries executed failed
	if errorCount > 0 {
		return fmt.Errorf("One or More Queries Failed")
	}

	return nil
}

//---------------------------------------------------------------------------------------

// Execute Query
func (qd *QueryDetails) ExecuteQuery(ctx context.Context, client *bigquery.Client, project string, dataset string, location string, disableQueryCache bool, delimiter string) {
	// Make Sure the Output File Path Exists
	path := filepath.Dir(qd.OutputFile)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, 0700); err != nil {
			qd.Error = fmt.Errorf("Failed to Create the Output File Path")
			return
		}
	}

	// Create and Configure Query
	q := client.Query(qd.SQL)
	q.DefaultProjectID = project
	q.DefaultDatasetID = dataset
	q.Location = location
	q.DisableQueryCache = disableQueryCache
	q.DryRun = false

	// Initiate the Query Job
	qd.QueryStartTime = time.Now()
	it, err := q.Read(ctx)
	qd.QueryEndTime = time.Now()
	if err != nil {
		qd.Error = err
		return
	}

	// Open the Output File
	f, err := os.Create(qd.OutputFile)
	if err != nil {
		qd.Error = fmt.Errorf("Failed to Open the Output File")
		return
	}
	defer f.Close()

	// Ready the CSV Writer and use a buffered io writer
	w := csv.NewWriter(bufio.NewWriter(f))
	w.Comma = rune(delimiter[0])
	defer w.Flush()

	var rl RowLoader
	var rowCount int64
	for {
		err := it.Next(&rl)
		if rowCount == 0 {
			qd.FirstRowReturnedTime = time.Now()
		}
		if err == iterator.Done {
			qd.AllRowsReturnedTime = time.Now()
			qd.TotalRowsReturned = rowCount
			break
		}
		if err != nil {
			qd.Error = err
			return
		}
		if err := w.Write(rl.Row); err != nil {
			qd.Error = fmt.Errorf("Failed Writing to the Output File")
			return
		}
		rowCount++
	}
}

//---------------------------------------------------------------------------------------

// Execute Dry Run Query
func (qd *QueryDetails) ExecuteDryRun(ctx context.Context, client *bigquery.Client, project string, dataset string, location string, disableQueryCache bool) {
	// Create and Configure Query
	q := client.Query(qd.SQL)
	q.DefaultProjectID = project
	q.DefaultDatasetID = dataset
	q.Location = location
	q.DisableQueryCache = disableQueryCache
	q.DryRun = true

	// Initiate the Query Job
	qd.QueryStartTime = time.Now()
	job, err := q.Run(ctx)
	if err != nil {
		qd.Error = err
		return
	}

	// Check the Last Status for Errors
	status := job.LastStatus()
	if err = status.Err(); err != nil {
		qd.Error = err
		return
	}
	qd.QueryEndTime = time.Now()
}

//---------------------------------------------------------------------------------------

// Output the Query Execution Statistics to the Log
func (qd *QueryDetails) LogExecuteQuery() {
	logger.Info().Int("Query Number", qd.Number).Msg("Query Execution")
	logger.Info().Str("Input File", qd.InputFile).Msg(indent)

	// Output Error Message if one exists, but nothing else
	if qd.Error != nil {
		logger.Error().Err(qd.Error).Msg(indent)
		return
	}

	logger.Info().Str("Output File", qd.OutputFile).Msg(indent)
	logger.Info().Time("Query Execution Start", qd.QueryStartTime).Msg(indent)
	logger.Info().Time("Query Execution End", qd.QueryEndTime).Msg(indent)
	logger.Info().TimeDiff("Execution Time (ms)", qd.QueryEndTime, qd.QueryStartTime).Msg(indent)
	logger.Info().Time("First Row Returned", qd.FirstRowReturnedTime).Msg(indent)
	logger.Info().Time("All Rows Returned", qd.AllRowsReturnedTime).Msg(indent)
	logger.Info().TimeDiff("Return Time (ms)", qd.AllRowsReturnedTime, qd.QueryEndTime).Msg(indent)
	logger.Info().Int64("Total Rows Returned", qd.TotalRowsReturned).Msg(indent)
}

//---------------------------------------------------------------------------------------

// Output the Query Dry Run Statistics to the Log
func (qd *QueryDetails) LogExecuteDryRun() {
	logger.Info().Int("Query Number", qd.Number).Msg("Query Dry Run")
	logger.Info().Str("Input File", qd.InputFile).Msg(indent)

	// Output Error Message if one exists, but nothing else
	if qd.Error != nil {
		logger.Error().Err(qd.Error).Msg(indent)
		return
	}

	logger.Info().Time("Query Execution Start", qd.QueryStartTime).Msg(indent)
	logger.Info().Time("Query Execution End", qd.QueryEndTime).Msg(indent)
	logger.Info().TimeDiff("Execution Time (ms)", qd.QueryEndTime, qd.QueryStartTime).Msg(indent)
}
