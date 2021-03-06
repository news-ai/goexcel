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
	customFields := make(map[string]bool, len(headers))

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

func xlsxToContactList(r *http.Request, file []byte, headers []string) ([]models.Contact, map[string]bool, error) {
	c := appengine.NewContext(r)

	xlsxFile, err := xlsx.OpenBinary(file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	if len(xlsxFile.Sheets) == 0 {
		err = errors.New("Sheet is empty")
		log.Errorf(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	sheet := xlsxFile.Sheets[0]

	if len(sheet.Rows) == 0 {
		err = errors.New("No rows in sheet")
		log.Errorf(c, "%v", err)
		return []models.Contact{}, map[string]bool{}, err
	}

	// Number of columns in sheet to compare
	numberOfColumns := len(sheet.Rows[0].Cells)
	startingPosition := 0

	// If number of columns is zero (edge case)
	if numberOfColumns == 0 {
		for i := 0; i < len(sheet.Rows); i++ {
			if len(sheet.Rows[i].Cells) > 0 {
				startingPosition = i
				numberOfColumns = len(sheet.Rows[i].Cells)
				break
			}
		}
	}

	if numberOfColumns != len(headers) {
		return []models.Contact{}, map[string]bool{}, errors.New("Number of headers does not match the ones for the sheet")
	}

	// Loop through all the rows
	// Extract information
	emptyContact := models.Contact{}
	contacts := []models.Contact{}
	for _, row := range sheet.Rows[startingPosition:] {
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
	customFields := getCustomFields(r, c, len(sheet.Rows[startingPosition].Cells), headers)

	return contacts, customFields, nil
}

func xlsxFileToExcelSheets(r *http.Request, file []byte) (Sheet, error) {
	c := appengine.NewContext(r)
	xlsxFile, err := xlsx.OpenBinary(file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return Sheet{}, err
	}

	if len(xlsxFile.Sheets) == 0 {
		err = errors.New("Sheet is empty")
		log.Errorf(c, "%v", err)
		return Sheet{}, err
	}

	sheetNames := Sheet{}

	for i := 0; i < len(xlsxFile.Sheets); i++ {
		sheetNames.Names = append(sheetNames.Names, xlsxFile.Sheets[i].Name)
	}

	return sheetNames, nil
}

func xlsxFileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	c := appengine.NewContext(r)
	xlsxFile, err := xlsx.OpenBinary(file)
	if err != nil {
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	if len(xlsxFile.Sheets) == 0 {
		err = errors.New("Sheet is empty")
		log.Errorf(c, "%v", err)
		return []Column{}, err
	}

	sheet := xlsxFile.Sheets[0]

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
	startingPosition := 0
	// If number of columns is zero (edge case)
	if numberOfColumns == 0 {
		for i := 0; i < len(sheet.Rows); i++ {
			if len(sheet.Rows[i].Cells) > 0 {
				startingPosition = i
				numberOfColumns = len(sheet.Rows[i].Cells)
				break
			}
		}
	}

	// Sometimes there are columns that are totally empty. Check if that is the case

	columns := make([]Column, numberOfColumns)
	for _, row := range sheet.Rows[startingPosition:numberOfRows] {
		// Skip row if it does not have the same amount of columns.
		// Might have a bug?
		if len(row.Cells) != numberOfColumns {
			continue
		}
		for currentColumn, cell := range row.Cells {
			cellName, _ := cell.String()
			columns[currentColumn].Rows = append(columns[currentColumn].Rows, strings.Trim(cellName, " "))
		}
	}

	return columns, nil
}
