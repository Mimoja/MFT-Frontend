package main

import (
	MFTCommon "MimojaFirmwareToolkit/pkg/Common"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func importOverviewHandler(c *gin.Context) {

	result := &ReportOverviewPage{
		Page: Page{Title: "Report",
			IsOkay: true},
	}

	numberOfResults := Bundle.Config.App.Frontend.ReportResults

	searchResult, err := Bundle.DB.ES.Search().
		Index("imports").
		Sort("ImportTime", false).
		From(0).Size(numberOfResults).
		Do(c.Request.Context())

	if err != nil {
		Bundle.Log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Could not query last reports")
		return
	}

	for _, hit := range searchResult.Hits.Hits {
		var importEntry MFTCommon.ImportEntry

		json.Unmarshal(*hit.Source, &importEntry)

		ref := NewImportRef(importEntry)
		result.LastReports = append(result.LastReports, ref)
	}

	display(c, "reportOverview", result)
}
