package main

import (
	"MimojaFirmwareToolkit/pkg/Common"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"net/http"
	"strconv"
)

type SearchPage struct {
	Page
	Error     string
	Time      string
	Hits      string
	Documents []SearchDocument
	Query     string
	Config    *MFTCommon.AppRunConfiguration
}

type SearchDocument struct {
	Hit         MFTCommon.FlashImage
	Parent      MFTCommon.ImportEntry
	DownloadURL string
}

func searchHandler(c *gin.Context) {
	result := &SearchPage{
		Page: Page{
			Title: "Search",
		},
		Config: &Bundle.Config.App,
	}

	query := c.Query("query")

	if query == "" {
		c.Status(http.StatusBadRequest)
		result.Error = "Query not specified"
		display(c, "search", result)
		return
	}
	skip := 0
	take := 10
	if i, err := strconv.Atoi(c.Query("skip")); err == nil {
		skip = i
	}
	if i, err := strconv.Atoi(c.Query("take")); err == nil {
		take = i
	}

	// Perform search
	esQuery := elastic.NewMultiMatchQuery(query, "MetaData.Vendor", "MetaData.Product", "MetaData.Version", "MetaData.Title", "MetaData.Description", "Tags")

	searchResult, err := Bundle.DB.ES.Search().
		Index("flashimages").
		Query(esQuery).
		From(skip).Size(take).
		Do(c.Request.Context())

	if err != nil {
		Bundle.Log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		return
	}

	result.Time = fmt.Sprintf("%d", searchResult.TookInMillis)
	result.Hits = fmt.Sprintf("%d", searchResult.Hits.TotalHits)

	docs := make([]SearchDocument, 0)
	for _, hit := range searchResult.Hits.Hits {
		var document SearchDocument
		var doc MFTCommon.FlashImage
		var importEntry MFTCommon.ImportEntry

		json.Unmarshal(*hit.Source, &doc)
		document.Hit = doc
		document.DownloadURL = "/file/" + doc.ID.GetID()

		exists, err, meta := Bundle.DB.Exists("imports", doc.MetaData.PackageID.GetID())

		if err != nil {
			Bundle.Log.Println("Could not get parent: ", err)
		} else if !exists {
			Bundle.Log.Println("Could not get parent: not exist")
		} else {
			json.Unmarshal(*meta.Source, &importEntry)
			document.Parent = importEntry
		}

		docs = append(docs, document)
	}
	result.Documents = docs
	result.Query = query

	display(c, "search", result)
}

/*
func searchEndpoint(c *gin.Context) {
	// Parse request

	query := c.Query("q")
	if query == "" {
		errorResponse(c, http.StatusBadRequest, "Query not specified")
		return
	}

	jsonOutput := false
	jsonOutput = c.Query("json") == "true"

	skip := 0
	take := 10
	if i, err := strconv.Atoi(c.Query("skip")); err == nil {
		skip = i
	}
	if i, err := strconv.Atoi(c.Query("take")); err == nil {
		take = i
	}
	// Perform search
	esQuery := elastic.NewMultiMatchQuery(query, "MetaData.Vendor", "MetaData.Product", "MetaData.Version", "MetaData.Title", "MetaData.Description", "Tags")

	result, err := Bundle.DB.ES.Search().
		Index("flashimages").
		Query(esQuery).
		From(skip).Size(take).
		Do(c.Request.Context())

	if err != nil {
		log.Println(err)
		errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		return
	}
	res := SearchResponse{
		Time:   fmt.Sprintf("%d", result.TookInMillis),
		Hits:   fmt.Sprintf("%d", result.Hits.TotalHits),
		Config: &Bundle.Config.App,
	}
	// Transform search results before returning them
	docs := make([]SearchDocument, 0)
	for _, hit := range result.Hits.Hits {
		var document SearchDocument
		var doc MFTCommon.FlashImage
		var importEntry MFTCommon.ImportEntry

		json.Unmarshal(*hit.Source, &doc)
		document.Hit = doc
		document.DownloadURL = "/file/" + doc.ID.GetID()

		exists, err, meta := Bundle.DB.Exists("imports", doc.MetaData.PackageID.GetID())

		if err != nil {
			log.Println("Could not get parent: ", err)
		} else if !exists {
			log.Println("Could not get parent: not exist")
		} else {
			json.Unmarshal(*meta.Source, &importEntry)
			document.Parent = importEntry
		}

		for _, cert := range doc.Certificates {
			exists, err, meta := Bundle.DB.Exists("certificates", cert)

			if err != nil {
				log.Println("Could not get cert: ", err)
			} else if !exists {
				log.Println("Could not get cert: not exist")
			} else {
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

				releaseDate, _ := time.Parse(time.RFC3339, doc.MetaData.ReleaseDate)

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
					continue
				}
				certDoc := CertificateDocument{
					Raw:            newCert,
					Valid:          validity_end.After(time.Now()),
					ValidAtRelease: validity_end.After(releaseDate),
					Subject:        subjectMap["common_name"][0],
					Issuer:         issuerMap["common_name"][0],
					Serial:         newCert["serial_number"].(string),
				}

				document.Certificates = append(document.Certificates, certDoc)

			}
		}

		docs = append(docs, document)
	}
	res.Documents = docs
	res.Query = query

	if jsonOutput {
		c.Header("Content-Type", "application/json")
		responseBytes, err := json.Marshal(res)
		if err != nil {
			log.Println("Could not execute template", err)
			errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		}
		c.Writer.Write(responseBytes)

	} else {
		tmpl := template.Must(template.ParseFiles("layout.html"))
		c.Header("Content-Type", "html")
		err = tmpl.Execute(c.Writer, res)

		if err != nil {
			log.Println("Could not execute template", err)
			errorResponse(c, http.StatusInternalServerError, "Something went wrong")
		}
	}

}
*/
