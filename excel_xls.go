package goexcel

import (
	"bytes"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/news-ai/tabulae/models"

	"google.golang.org/appengine"

	"golang.org/x/net/context"

	"github.com/extrame/xls"
)

func xlsRowToContact(r *http.Request, c context.Context, numberOfColumns int, workbook *xls.WorkBook, singleRow *xls.Row, headers []string) (models.Contact, error) {
	var (
		contact       models.Contact
		employers     []int64
		pastEmployers []int64
		customFields  []models.CustomContactField
	)

	for x := 0; x < numberOfColumns; x++ {
		columnName := headers[x]
		currentRow := singleRow.Cols[uint16(x)]
		cellName := ""
		if currentRow != nil {
			cellName = singleRow.Cols[uint16(x)].String(workbook)[0]
		}
		rowToContact(r, c, columnName, cellName, &contact, &employers, &pastEmployers, &customFields)
	}

	contact.CustomFields = customFields
	contact.Employers = employers
	contact.PastEmployers = pastEmployers
	return contact, nil
}

func XlsToContactList(r *http.Request, file []byte, headers []string) ([]models.Contact, map[string]bool, error) {
	c := appengine.NewContext(r)

	readerFile := bytes.NewReader(file)
	workbook, err := xls.OpenReader(readerFile, "utf-8")
	if err != nil {
		return []models.Contact{}, map[string]bool{}, err
	}

	sheet := workbook.GetSheet(0)
	if sheet == nil {
		return []models.Contact{}, map[string]bool{}, errors.New("Sheet is empty")
	}

	// Number of columns in sheet to compare
	numberOfColumns := len(sheet.Rows[0].Cols)
	if numberOfColumns != len(headers) {
		return []models.Contact{}, map[string]bool{}, errors.New("Number of headers does not match the ones for the sheet")
	}

	// Loop through all the rows
	// Extract information
	emptyContact := models.Contact{}
	contacts := []models.Contact{}

	for i := 0; i < int(sheet.MaxRow); i++ {
		row := sheet.Rows[uint16(i)]
		contact, err := xlsRowToContact(r, c, numberOfColumns, workbook, row, headers)
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

func XlsFileToExcelHeader(r *http.Request, file []byte) ([]Column, error) {
	readerFile := bytes.NewReader(file)
	workbook, err := xls.OpenReader(readerFile, "utf-8")
	if err != nil {
		return []Column{}, err
	}

	sheet := workbook.GetSheet(0)
	if sheet == nil {
		return []Column{}, errors.New("Sheet is empty")
	}

	// Number of rows to consider
	numberOfRows := 15
	if int(sheet.MaxRow) < numberOfRows+1 {
		numberOfRows = int(sheet.MaxRow)
	}

	numberOfColumns := len(sheet.Rows[0].Cols)
	columns := make([]Column, numberOfColumns)

	for i := 0; i <= numberOfRows; i++ {
		row := sheet.Rows[uint16(i)]
		for x := 0; x < numberOfColumns; x++ {
			currentRow := row.Cols[uint16(x)]
			cellName := ""
			if currentRow != nil {
				cellName = row.Cols[uint16(x)].String(workbook)[0]
			}
			columns[x].Rows = append(columns[x].Rows, strings.Trim(cellName, " "))
		}
	}

	return columns, nil
}
