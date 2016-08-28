package goexcel

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func CsvFileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	c := appengine.NewContext(r)

	readerFile := bytes.NewReader(file)
	incomingRecords := csv.NewReader(readerFile)
	records, err := incomingRecords.ReadAll()
	if err != nil {
		log.Errorf(c, "%v", err)
	}

	log.Infof(c, "%v", records)

	numberOfRows := 15
	if len(records) < numberOfRows+1 {
		numberOfRows = len(records)
	}

	numberOfColumns := len(records[0])
	columns := make([]Column, numberOfColumns)

	for _, row := range records[0:numberOfRows] {
		for currentColumn, cell := range row {
			columns[currentColumn].Rows = append(columns[currentColumn].Rows, strings.Trim(cell, " "))
		}
	}

	return columns, nil
}
