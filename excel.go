package goexcel

import (
	"errors"
	"net/http"

	"google.golang.org/appengine/log"

	"golang.org/x/net/context"

	"github.com/news-ai/tabulae/controllers"
	"github.com/news-ai/tabulae/models"
)

var nonCustomHeaders = map[string]bool{
	"firstname":     true,
	"lastname":      true,
	"email":         true,
	"employers":     true,
	"pastemployers": true,
	"notes":         true,
	"linkedin":      true,
	"twitter":       true,
	"instagram":     true,
	"website":       true,
	"blog":          true,
	"location":      true,
	"phonenumber":   true,
}

type Column struct {
	Rows []string `json:"rows"`
}

type Sheet struct {
	Names []string `json:"names"`
}

func customOrNative(columnName string) bool {
	if _, ok := nonCustomHeaders[columnName]; ok {
		return true
	}
	return false
}

func getCustomFields(r *http.Request, c context.Context, numberOfColumns int, headers []string) map[string]bool {
	customFields := make(map[string]bool, len(headers))

	for x := 0; x < numberOfColumns; x++ {
		columnName := headers[x]
		if !customOrNative(columnName) {
			customFields[columnName] = true
		}
	}
	return customFields
}

func rowToContact(r *http.Request, c context.Context, columnName string, cellName string, contact *models.Contact, employers *[]int64, pastEmployers *[]int64, customFields *[]models.CustomContactField) {
	if columnName != "ignore_column" {
		if customOrNative(columnName) {
			switch columnName {
			case "firstname":
				contact.FirstName = cellName
			case "lastname":
				contact.LastName = cellName
			case "email":
				contact.Email = cellName
			case "notes":
				contact.Notes = cellName
			case "employers":
				if cellName != "" {
					singleEmployer, err := controllers.UploadFindOrCreatePublication(c, r, cellName, "")
					if err != nil {
						log.Errorf(c, "employers error: %v", cellName, err)
					}
					*employers = append(*employers, singleEmployer.Id)
				}
			case "pastemployers":
				if cellName != "" {
					singleEmployer, err := controllers.UploadFindOrCreatePublication(c, r, cellName, "")
					if err != nil {
						log.Errorf(c, "past employers error: %v", cellName, err)
					}
					*pastEmployers = append(*pastEmployers, singleEmployer.Id)
				}
			case "linkedin":
				contact.LinkedIn = cellName
			case "twitter":
				contact.Twitter = cellName
			case "instagram":
				contact.Instagram = cellName
			case "website":
				contact.Website = cellName
			case "blog":
				contact.Blog = cellName
			case "location":
				contact.Location = cellName
			case "phonenumber":
				contact.PhoneNumber = cellName
			}
		} else {
			var customField models.CustomContactField
			customField.Name = columnName
			customField.Value = cellName
			*customFields = append(*customFields, customField)
		}
	}
}

func FileToExcelSheets(c context.Context, r *http.Request, file []byte, contentType string) (Sheet, error) {
	if contentType == "application/vnd.ms-excel" {
		log.Infof(c, "%v", contentType)
		return Sheet{}, nil
	} else if contentType == "text/csv" {
		log.Infof(c, "%v", contentType)
		return Sheet{}, nil
	}
	return xlsxFileToExcelSheets(r, file)
}

func FileToExcelHeader(c context.Context, r *http.Request, file []byte, contentType string) ([]Column, error) {
	if contentType == "application/vnd.ms-excel" {
		log.Infof(c, "%v", contentType)
		return xlsFileToExcelHeader(r, file)
	} else if contentType == "text/csv" {
		log.Infof(c, "%v", contentType)
		return csvFileToExcelHeader(r, file)
	}
	return xlsxFileToExcelHeader(r, file)
}

func HeadersToListModel(c context.Context, r *http.Request, file []byte, headers []string, contentType string) ([]models.Contact, map[string]bool, error) {
	contacts := []models.Contact{}
	var customFields map[string]bool
	err := errors.New("")

	if contentType == "application/vnd.ms-excel" {
		log.Infof(c, "%v", contentType)
		contacts, customFields, err = xlsToContactList(r, file, headers)
		if err != nil {
			return []models.Contact{}, customFields, err
		}
	} else if contentType == "text/csv" {
		log.Infof(c, "%v", contentType)
		contacts, customFields, err = csvToContactList(r, file, headers)
		if err != nil {
			return []models.Contact{}, customFields, err
		}
	} else {
		contacts, customFields, err = xlsxToContactList(r, file, headers)
		if err != nil {
			return []models.Contact{}, customFields, err
		}
	}

	return contacts, customFields, err
}
