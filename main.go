package main

import (
	"MimojaFirmwareToolkit/pkg/Common"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

const host = ""
const port = 9080

var Bundle MFTCommon.AppBundle

func errorResponse(c *gin.Context, code int, err string) {
	c.String(code, "error %d: %s", code, err)
}

const templateFolder = "templates"

func getAllTemplates() []string {
	templateFiles := []string{}
	files, err := ioutil.ReadDir(templateFolder)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		templateFiles = append(templateFiles, path.Join(templateFolder, f.Name()))
	}
	fmt.Println(templateFiles)
	return templateFiles
}

var templates = template.Must(template.New("dummy").Funcs(sprig.FuncMap()).ParseFiles(getAllTemplates()...))

func display(c *gin.Context, tmpl string, data interface{}) {
	c.Header("Content-Type", "html")

	//TODO remove for production
	//templates = template.Must(template.ParseFiles(getAllTemplates()...))

	err := templates.ExecuteTemplate(c.Writer, tmpl, data)
	if err != nil {
		Bundle.Log.WithError(err).Error("Template error: ", err)
		errorResponse(c, http.StatusInternalServerError, "Could not activate template")
	}
}

func mainHandler(c *gin.Context) {
	display(c, "main", &Page{Title: "MimojaFirmwareToolkit"})

}

func aboutHandler(c *gin.Context) {
	display(c, "about", &Page{Title: "About"})
}

func fileHandler(c *gin.Context) {
	if !Bundle.Config.App.Frontend.DownloadEnabled {
		errorResponse(c, http.StatusNotFound, "Not found")
		return
	}

	query := c.Params.ByName("file")
	if query == "" {
		errorResponse(c, http.StatusBadRequest, "File not specified")
		return
	}

	object, err := Bundle.Storage.GetFile(query)
	if err != nil {
		errorResponse(c, 404, "Could not fetch file")
		return
	}
	defer object.Close()

	stats, err := object.Stat()
	if err != nil {
		errorResponse(c, 404, "Could not fetch file: Not exist")
		return
	}

	if stats.Size == 0 {
		errorResponse(c, 404, "File is empty")
		return
	}

	c.Header("Content-Type", "application/octet-stream")

	if _, err := io.Copy(c.Writer, object); err != nil {
		errorResponse(c, 500, "Internal error: "+err.Error())
	}
}

func main() {
	Bundle = MFTCommon.Init("FlashCatalog")

	hostAndPort := fmt.Sprintf("%s:%d", host, port)
	Bundle.Log.Infof("Starting http server on %s\n", hostAndPort)

	r := gin.Default()
	r.GET("/", mainHandler)
	r.Use(static.Serve("/static/", static.LocalFile("./static", false)))
	r.GET("/report", importOverviewHandler)
	r.GET("/report/:reportID", reportIDHandler)
	r.GET("/library", libraryHandler)
	r.GET("/contribute", contributeHandler)
	r.GET("/about", aboutHandler)
	r.GET("/search", searchHandler)
	r.POST("/upload", uploadHandler)
	r.GET("/file/:file", fileHandler)

	if err := r.Run(":9080"); err != nil {
		log.Fatal(err)
	}
}
