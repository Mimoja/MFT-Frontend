package main

import (
	"encoding/json"
	"github.com/Mimoja/MFT-Common"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"strings"
	"time"
)

func reportIDHandler(c *gin.Context) {
	var jsonOutput bool
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

	jsonOutput = c.Query("json") == "true"

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
		err = json.Unmarshal(sourceBytes, &page.Data.Import)
		if err != nil {
			Bundle.Log.WithError(err).WithField("payload", string(sourceBytes)).Warnf("Could unmarshall old entry from elastic: %v", err)
			errorResponse(c, http.StatusBadRequest, "Report not found")
		}
	}

	for _, b := range page.Data.Import.Contents {
		exists, err, value := Bundle.DB.Exists("flashimages", b.ID.GetID())
		if err != nil {
			logrus.Info("Could not query bug %s from elastic: ", err, b)
			continue
		}
		if !exists {
			continue
		}
		var flashimage MFTCommon.FlashImage
		sourceBytes, _ := value.Source.MarshalJSON()
		err = json.Unmarshal(sourceBytes, &flashimage)
		if err != nil {
			logrus.Info("Could not unmarshall flashimage ", err)
			continue
		}
		flashDocument := FlashDocument{FlashImage: flashimage}

		for _, cert := range flashimage.Certificates {
			exists, err, meta := Bundle.DB.Exists("certificates", cert)

			if err != nil {
				log.Println("Could not get cert: ", err)
			} else if !exists {
				log.Println("Could not get cert: not exist")
			}
			var newCert map[string]interface{}
			certBytes, err := meta.Source.MarshalJSON()
			err = json.Unmarshal(certBytes, &newCert)
			if err != nil {
				log.Println("Could not unmarshall cert: ", err)
				continue
			}

			validity := newCert["validity"].(map[string]interface{})
			validity_end, _ := time.Parse(time.RFC3339, validity["end"].(string))
			//validity_start,_ := time.Parse(time.RFC3339,validity["start"].(string))

			subjectArrayJSON, _ := json.Marshal(newCert["subject"])
			var subjectMap map[string][]string
			err = json.Unmarshal(subjectArrayJSON, &subjectMap)
			if err != nil {
				log.Println("Could not unmarshall cert: ", err)
				continue
			}

			issuerArrayJSON, _ := json.Marshal(newCert["issuer"])
			var issuerMap map[string][]string
			err = json.Unmarshal(issuerArrayJSON, &issuerMap)
			if err != nil {
				log.Println("Could not unmarshall cert: ", err)
				continue
			}
			if len(subjectMap) == 0 || len(issuerMap) == 0 {
				log.Println("Cert is empty")
				continue
			}

			certDoc := CertificateDocument{
				Raw:     newCert,
				Valid:   validity_end.After(time.Now()),
				Subject: subjectMap["common_name"],
				Issuer:  issuerMap["common_name"],
				Serial:  newCert["serial_number"].(string),
			}

			flashDocument.Certificates = append(flashDocument.Certificates, certDoc)

		}
		if flashDocument.FlashImage.EFIBlob != "" {
			flashDocument.FlashImage.EFIBlob = strings.ReplaceAll(flashDocument.FlashImage.EFIBlob, "\n","")
			flashDocument.FlashImage.EFIBlob = strings.ReplaceAll(flashDocument.FlashImage.EFIBlob, "\t","")
			err = json.Unmarshal([]byte(flashDocument.FlashImage.EFIBlob), &flashDocument.EFI)

			if err != nil {
				log.Println("Could not unmarshall EFI: ", err)
			}
		}

		page.Data.FlashImages = append(page.Data.FlashImages, flashDocument)
	}
	if jsonOutput {
		c.Header("Content-Type", "application/json")
		responseBytes, err := json.MarshalIndent(page.Data, "", " ")
		if err != nil {
			log.Println("Could not execute template", err)
			errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		}
		c.Writer.Write(responseBytes)
		return;
	}

	display(c, "report", page)
}
