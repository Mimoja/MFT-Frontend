package main

import (
	"MimojaFirmwareToolkit/pkg/Common"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

func uploadHandler(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()

	uploadedBytes, _ := ioutil.ReadAll(src)

	ID := MFTCommon.GenerateID(uploadedBytes)

	err = Bundle.Storage.StoreBytes(uploadedBytes, ID.GetID())

	if err != nil {
		Bundle.Log.WithError(err).Error("Could not store file to storage! ", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Could not handle file. Please try again"))
	}

	entry := MFTCommon.UserUpload{
		MetaData: MFTCommon.DownloadEntry{
			DownloadURL:  "userupload/" + file.Filename,
			DownloadPath: "userupload/" + file.Filename,
			PackageID:    ID,
		},
		UploadIP:   c.ClientIP(),
		UploadTime: time.Now().Format(time.RFC3339),
	}

	idString := ID.GetID()
	Bundle.DB.StoreElement("useruploads", nil, entry, &idString)

	err = Bundle.MessageQueue.DownloadedQueue.MarshalAndSend(entry.MetaData)

	if err != nil {
		Bundle.Log.WithError(err).Error("Could not store file to storage! ", err)
		c.String(http.StatusInternalServerError, fmt.Sprintf("Could not handle file. Please try again"))
	}

	c.Redirect(http.StatusFound, "/report/"+ID.GetID())
}
