package main

import (
	"github.com/Mimoja/MFT-Common"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func reportIDHandler(c *gin.Context) {
	page := &ReportPage{
		Page: Page{
			Title:  "Report",
			IsOkay: true,
		},
		Config: &Bundle.Config.App,
	}

	query := c.Params.ByName("reportID")
	if query == "" {
		errorResponse(c, http.StatusBadRequest, "Report not specified")
		return
	}
	found, err, result := Bundle.DB.Exists("imports", query)

	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Something went wrong")
		return
	}

	if !found {
		page.Page.Error = "Report was not found or is not ready yet. Please try again in a moment"
		page.Page.IsOkay = false
		display(c, "report", page)
		return
	}

	sourceBytes, err := result.Source.MarshalJSON()
	if err != nil {
		Bundle.Log.WithError(err).Info("Could not get old entry from elastic: %v", err)
		errorResponse(c, http.StatusBadRequest, "Report not found")
	} else {
		err = json.Unmarshal(sourceBytes, &page.Import)
		if err != nil {
			Bundle.Log.WithError(err).WithField("payload", string(sourceBytes)).Warnf("Could unmarshall old entry from elastic: %v", err)
			errorResponse(c, http.StatusBadRequest, "Report not found")

		}

	}

	//contenttree
	basePath := ""
	for _, c := range page.Import.Contents {
		for _, tag := range c.Tags {
			if tag == "DOWNLOAD" {
				basePath = filepath.Dir(c.Path)
			}
		}
	}
	if basePath == "." {
		basePath = ""
	}

	fmt.Println("Basepath is: ", basePath)
	contentTree := make(map[string]interface{})
	for _, c := range page.Import.Contents {
		c.Path = c.Path[len(basePath):]
		c.Path = strings.Trim(c.Path, "./")
		contentTree[c.Path] = c
	}

	for _, b := range page.Import.Contents {

		exists, err, value := Bundle.DB.Exists("flashimages", b.ID.GetID())
		if err != nil {
			logrus.Info("Could not query bug %s from elastic: ", err, b)
			continue
		}
		if exists {
			var flashimage MFTCommon.FlashImage
			sourceBytes, _ := value.Source.MarshalJSON()
			err = json.Unmarshal(sourceBytes, &flashimage)
			page.FlashImages = append(page.FlashImages, flashimage)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(contentTree)

	display(c, "report", page)
}
