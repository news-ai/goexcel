package goexcel

import (
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
}

type Column struct {
	Rows []string `json:"rows"`
}

func customOrNative(columnName string) bool {
	if _, ok := nonCustomHeaders[columnName]; ok {
		return true
	}
	return false
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
				singleEmployer, err := controllers.FindOrCreatePublication(c, r, cellName)
				if err != nil {
					log.Errorf(c, "employers error: %v", cellName, err)
				}
				*employers = append(*employers, singleEmployer.Id)
			case "pastemployers":
				singleEmployer, err := controllers.FindOrCreatePublication(c, r, cellName)
				if err != nil {
					log.Errorf(c, "past employers error: %v", cellName, err)
				}
				*pastEmployers = append(*pastEmployers, singleEmployer.Id)
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
			}
		} else {
			var customField models.CustomContactField
			customField.Name = columnName
			customField.Value = cellName
			*customFields = append(*customFields, customField)
		}
	}
}
