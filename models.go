package main

import (
	MFTCommon "MimojaFirmwareToolkit/pkg/Common"
	"path/filepath"
	"time"
)

type Page struct {
	Title  string
	Error  string
	IsOkay bool
}

type ReportOverviewPage struct {
	Page
	LastReports []ImportRef
}

type ImportRef struct {
	ImportTime string
	ID         string
	Name       string
}

/*
type ImportEntry struct {
	ImportDataDefinition string         `json:",omitempty"`
	MetaData             DownloadEntry  `json:",omitempty"`
	Contents             []StorageEntry `json:",omitempty"`
	ImportTime           string         `json:",omitempty"`
	Success              bool           `json:",omitempty"`
}
*/

func NewImportRef(importEntry MFTCommon.ImportEntry) ImportRef {

	name := filepath.Base(importEntry.MetaData.DownloadPath)

	if name == "." {
		name = ""
	}

	importTime, _ := time.Parse(time.RFC3339, importEntry.ImportTime)
	return ImportRef{
		ImportTime: importTime.Format("02.01.2006 15:04:05"),
		ID:         importEntry.MetaData.PackageID.GetID(),
		Name:       name,
	}
}

type ReportPage struct {
	Page
	UploadMeta  ImportRef
	Import      MFTCommon.ImportEntry
	Config      *MFTCommon.AppRunConfiguration
	FlashImages []MFTCommon.FlashImage
}
