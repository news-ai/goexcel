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
	var (
		contact       models.Contact
		employers     []int64
		pastEmployers []int64
		customFields  []models.CustomContactField
	)

	for x := 0; x < len(singleRow); x++ {
		columnName := headers[x]
		cellName := singleRow[x]
		rowToContact(r, c, columnName, cellName, &contact, &employers, &pastEmployers, &customFields)
	}

	contact.CustomFields = customFields
	contact.Employers = employers
	contact.PastEmployers = pastEmployers
	return contact, nil
}

func csvToContactList(r *http.Request, file []byte, headers []string) ([]models.Contact, map[string]bool, error) {
	c := appengine.NewContext(r)

	readerFile := bytes.NewReader(file)
	incomingRecords := csv.NewReader(readerFile)
	records, err := incomingRecords.ReadAll()
	if err != nil {
		log.Infof(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	// Number of columns in sheet to compare
	numberOfColumns := len(records[0])
	if numberOfColumns != len(headers) {
		err := errors.New("Number of headers does not match the ones for the sheet")
		log.Infof(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	// Loop through all the rows
	// Extract information
	emptyContact := models.Contact{}
	contacts := []models.Contact{}

	for i := 0; i < len(records); i++ {
		contact, err := csvRowToContact(r, c, records[i], headers)
		if err != nil {
			log.Infof(c, "%v", err)
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

func csvFileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	c := appengine.NewContext(r)

	readerFile := bytes.NewReader(file)
	incomingRecords := csv.NewReader(readerFile)
	records, err := incomingRecords.ReadAll()
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

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
