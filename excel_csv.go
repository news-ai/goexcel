package goexcel

import (
	"bytes"
	"encoding/csv"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/models"
)

func csvGetCustomFields(r *http.Request, c context.Context, numberOfColumns int, headers []string) map[string]bool {
	customFields := make(map[string]bool, len(headers))

	for x := 0; x < numberOfColumns; x++ {
		columnName := headers[x]
		if !customOrNative(columnName) {
			customFields[columnName] = true
		}
	}
	return customFields
}

func csvRowToContact(r *http.Request, c context.Context, singleRow []string, headers []string) (models.Contact, error) {

}

func CsvToContactList(r *http.Request, file []byte, headers []string) ([]models.Contact, map[string]bool, error) {
	c := appengine.NewContext(r)

	readerFile := bytes.NewReader(file)
	incomingRecords := csv.NewReader(readerFile)
	records, err := incomingRecords.ReadAll()

	// Number of columns in sheet to compare
	numberOfColumns := len(records[0])
	if numberOfColumns != len(headers) {
		return []models.Contact{}, map[string]bool{}, errors.New("Number of headers does not match the ones for the sheet")
	}

	// Loop through all the rows
	// Extract information
	emptyContact := models.Contact{}
	contacts := []models.Contact{}

	for i := 0; i < len(records); i++ {
		contact, err := csvRowToContact(r, c, records[0], headers)
		if err != nil {
			return []models.Contact{}, map[string]bool{}, err
		}

		// To get rid of empty contacts. We don't want to create empty contacts.
		if !reflect.DeepEqual(emptyContact, contact) {
			contacts = append(contacts, contact)
		}
	}

	// Get custom fields
	customFields := getCustomFields(r, c, numberOfColumns, headers)

	return contacts, customFields, nil
}

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
