package main

import (
	"encoding/json"
	"fmt"
	"github.com/Mimoja/MFT-Common"
	"github.com/gin-gonic/gin"
	"net/http"
)

func rescanHandler(c *gin.Context) {
	var importEntry MFTCommon.ImportEntry

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
		c.Redirect(http.StatusFound, "/")
		return
	}

	sourceBytes, err := result.Source.MarshalJSON()
	if err != nil {
		Bundle.Log.WithError(err).Info("Could not get old entry from elastic: %v", err)
		errorResponse(c, http.StatusBadRequest, "Report not found")
	} else {
		err = json.Unmarshal(sourceBytes, &importEntry)
		if err != nil {
			Bundle.Log.WithError(err).WithField("payload", string(sourceBytes)).Warnf("Could unmarshall old entry from elastic: %v", err)
			errorResponse(c, http.StatusBadRequest, "Report not found")
		}
	}

	err = Bundle.MessageQueue.DownloadedQueue.MarshalAndSend(MFTCommon.DownloadWrapper{importEntry.MetaData, true})

	if err != nil {
		Bundle.Log.WithError(err).Error("Could not store file to storage! ", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Could not handle file. Please try again"))
	}

	c.Redirect(http.StatusFound, "/report/"+importEntry.MetaData.PackageID.GetID())

}
