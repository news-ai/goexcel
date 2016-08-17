package goexcel

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/models"

	"github.com/tealeg/xlsx"
)

func xlsxGetCustomFields(r *http.Request, c context.Context, numberOfColumns int, headers []string) map[string]bool {
	var customFields map[string]bool

	for i := 0; i < numberOfColumns; i++ {
		columnName := headers[i]
		if !customOrNative(columnName) {
			customFields[columnName] = true
		}
	}
	return customFields
}

func xlsxRowToContact(r *http.Request, c context.Context, singleRow *xlsx.Row, headers []string) (models.Contact, error) {
	var (
		contact       models.Contact
		employers     []int64
		pastEmployers []int64
		customFields  []models.CustomContactField
	)

	for columnIndex, cell := range singleRow.Cells {
		columnName := headers[columnIndex]
		cellName, _ := cell.String()
		rowToContact(r, c, columnName, cellName, &contact, &employers, &pastEmployers, &customFields)
	}

	contact.CustomFields = customFields
	contact.Employers = employers
	contact.PastEmployers = pastEmployers
	return contact, nil
}

func XlsxToContactList(r *http.Request, file []byte, headers []string, mediaListid int64) ([]models.Contact, map[string]bool, error) {
	c := appengine.NewContext(r)

	xlFile, err := xlsx.OpenBinary(file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	if len(xlFile.Sheets) == 0 {
		err = errors.New("Sheet is empty")
		log.Errorf(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	sheet := xlFile.Sheets[0]

	if len(sheet.Rows) == 0 {
		err = errors.New("No rows in sheet")
		log.Errorf(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	// Number of columns in sheet to compare
	numberOfColumns := len(sheet.Rows[0].Cells)
	if numberOfColumns != len(headers) {
		return []models.Contact{}, map[string]bool{}, errors.New("Number of headers does not match the ones for the sheet")
	}

	// Loop through all the rows
	// Extract information
	emptyContact := models.Contact{}
	contacts := []models.Contact{}
	for _, row := range sheet.Rows {
		contact, err := xlsxRowToContact(r, c, row, headers)
		if err != nil {
			return []models.Contact{}, map[string]bool{}, err
		}

		// To get rid of empty contacts. We don't want to create empty contacts.
		if !reflect.DeepEqual(emptyContact, contact) {
			contacts = append(contacts, contact)
		}
	}

	// Get custom fields
	customFields := xlsxGetCustomFields(r, c, len(sheet.Rows[0].Cells), headers)

	return contacts, customFields, nil
}

func XlsxFileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	c := appengine.NewContext(r)
	xlFile, err := xlsx.OpenBinary(file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	if len(xlFile.Sheets) == 0 {
		err = errors.New("Sheet is empty")
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	sheet := xlFile.Sheets[0]

	if len(sheet.Rows) == 0 {
		err = errors.New("No rows in sheet")
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	// Number of rows to consider
	numberOfRows := 15
	if len(sheet.Rows) < numberOfRows+1 {
		numberOfRows = len(sheet.Rows)
	}

	numberOfColumns := len(sheet.Rows[0].Cells)
	columns := make([]Column, numberOfColumns)

	for _, row := range sheet.Rows[0:numberOfRows] {
		for currentColumn, cell := range row.Cells {
			cellName, _ := cell.String()
			columns[currentColumn].Rows = append(columns[currentColumn].Rows, strings.Trim(cellName, " "))
		}
	}

	return columns, nil
}
