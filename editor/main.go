// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/redbluescreen/worldlangedit/lib"
)

var win *walk.MainWindow
var statusBar *walk.StatusBarItem
var logStatus *walk.StatusBarItem
var table *walk.TableView
var saveButton *walk.Action
var addButton *walk.Action
var searchEdit *walk.LineEdit

var labelsFile *lib.LangFile
var langFile *lib.LangFile
var langFilePath string
var labelsFilePath string
var labelsEdited bool
var tableEntries []TableEntry

type TableEntry struct {
	Hash        uint32
	IsNew       bool
	Label       string
	Translation string
}

func UpdateShownTableEntries() {
	search := strings.ToUpper(searchEdit.Text())
	if len(search) == 0 {
		table.SetModel(tableEntries)
	} else {
		shownEntries := make([]TableEntry, 0)
		for _, entry := range tableEntries {
			c1 := strings.ToUpper(entry.Label)
			c2 := strings.ToUpper(entry.Translation)
			if strings.Contains(c1, search) || strings.Contains(c2, search) {
				shownEntries = append(shownEntries, entry)
			}
		}
		table.SetModel(shownEntries)
	}
}

func toolOpenTriggered() {
	d := &walk.FileDialog{
		Filter: "Language files|*_Global.bin",
	}
	d.ShowOpen(win)
	langFilePath = d.FilePath
	enc, err := ioutil.ReadFile(d.FilePath)
	if err != nil {
		walk.MsgBox(win, "Error", "Failed to open language file", walk.MsgBoxIconError)
		return
	}
	langFile = lib.ParseFile(enc)
	labelsFilePath = path.Join(path.Dir(d.FilePath), "Labels_Global.bin")
	enc, err = ioutil.ReadFile(labelsFilePath)
	if err != nil {
		walk.MsgBox(win, "Error", "Failed to open Labels_Global.bin", walk.MsgBoxIconError)
		return
	}
	labelsFile = lib.ParseFile(enc)
	entriesMap := make(map[uint32]TableEntry)
	for _, entry := range labelsFile.Entries {
		sct := entriesMap[entry.Hash]
		sct.Label = entry.String
		entriesMap[entry.Hash] = sct
	}
	for _, entry := range langFile.Entries {
		sct := entriesMap[entry.Hash]
		sct.Translation = entry.String
		entriesMap[entry.Hash] = sct
	}
	entries := make([]TableEntry, len(entriesMap))
	i := 0
	for h, e := range entriesMap {
		e.Hash = h
		entries[i] = e
		i++
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Hash < entries[j].Hash })
	tableEntries = entries
	UpdateShownTableEntries()
	logStatus.SetText("File opened")
	statusBar.SetText(d.FilePath)
	saveButton.SetEnabled(true)
	addButton.SetEnabled(true)
}

func tableItemActivated() {
	entry := table.Model().([]TableEntry)[table.CurrentIndex()]
	var dlg *walk.Dialog
	var db *walk.DataBinder
	var labelEdit *walk.LineEdit
	var hashEdit *walk.NumberEdit
	origHash := entry.Hash
	Dialog{
		AssignTo: &dlg,
		Title:    "Edit translation",
		MinSize:  Size{500, 300},
		DataBinder: DataBinder{
			AssignTo:   &db,
			Name:       "entry",
			DataSource: &entry,
		},
		Layout: VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "Hash:",
					},
					NumberEdit{
						AssignTo: &hashEdit,
						Value:    Bind("Hash"),
						// TextAlignment: AlignNear,
						// Decimals:      0,
						ReadOnly: true,
					},
					Label{
						Text: "Label:",
					},
					LineEdit{
						AssignTo: &labelEdit,
						Text:     Bind("Label"),
						OnTextChanged: func() {
							t := labelEdit.Text()
							h := lib.BinHash(t)
							hashEdit.SetValue(float64(h))
						},
						ReadOnly: true,
					},
					Label{
						Text: "Translation:",
					},
					TextEdit{
						Text: Bind("Translation"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Save",
						OnClicked: func() {
							db.Submit()

							for i, e := range tableEntries {
								if e.Hash == origHash {
									tableEntries[i] = entry
									break
								}
							}
							UpdateShownTableEntries()
							for i, e := range langFile.Entries {
								if e.Hash == origHash {
									langFile.Entries[i].Hash = entry.Hash
									langFile.Entries[i].String = entry.Translation
									break
								}
							}
							for i, e := range labelsFile.Entries {
								if e.Hash == origHash {
									labelsFile.Entries[i].Hash = entry.Hash
									labelsFile.Entries[i].String = entry.Label
									break
								}
							}

							dlg.Accept()
						},
					},
				},
			},
		},
	}.Run(win)
}

func BackupFileIfNeeded(path string) {
	if _, err := os.Stat(path + ".bak"); os.IsNotExist(err) {
		err = os.Rename(path, path+".bak")
		fmt.Println(err)
	}
}

func toolSaveTriggered() {
	f := lib.SaveFile(langFile, labelsFile, true)
	BackupFileIfNeeded(langFilePath)
	ioutil.WriteFile(langFilePath, f, 666)
	if labelsEdited {
		f = lib.SaveFile(labelsFile, labelsFile, true)
		BackupFileIfNeeded(labelsFilePath)
		ioutil.WriteFile(labelsFilePath, f, 666)
	}
	logStatus.SetText("File saved")
}

func toolAddTriggered() {
	entry := TableEntry{}
	var dlg *walk.Dialog
	var db *walk.DataBinder
	var labelEdit *walk.LineEdit
	var hashEdit *walk.NumberEdit
	Dialog{
		AssignTo: &dlg,
		Title:    "Add translation",
		MinSize:  Size{500, 300},
		DataBinder: DataBinder{
			AssignTo:   &db,
			Name:       "entry",
			DataSource: &entry,
		},
		Layout: VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{
						Text: "Hash:",
					},
					NumberEdit{
						AssignTo: &hashEdit,
						Value:    Bind("Hash"),
						// TextAlignment: AlignNear,
						// Decimals:      0,
						ReadOnly: true,
					},
					Label{
						Text: "Label:",
					},
					LineEdit{
						AssignTo: &labelEdit,
						Text:     Bind("Label"),
						OnTextChanged: func() {
							t := labelEdit.Text()
							h := lib.BinHash(t)
							hashEdit.SetValue(float64(h))
						},
					},
					Label{
						Text: "Translation:",
					},
					TextEdit{
						Text: Bind("Translation"),
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Save",
						OnClicked: func() {
							db.Submit()

							tableEntries = append(tableEntries, entry)
							UpdateShownTableEntries()
							langFile.Entries = append(langFile.Entries, lib.LangFileEntry{
								Hash:   entry.Hash,
								String: entry.Translation,
							})
							labelsFile.Entries = append(labelsFile.Entries, lib.LangFileEntry{
								Hash:   entry.Hash,
								String: entry.Label,
							})
							labelsEdited = true

							dlg.Accept()
						},
					},
				},
			},
		},
	}.Run(win)
}

func main() {
	MainWindow{
		AssignTo: &win,
		Title:    "WorldLangEdit by redbluescreen",
		MinSize:  Size{800, 600},
		Layout: VBox{
			MarginsZero: true,
			SpacingZero: true,
		},
		Children: []Widget{
			TableView{
				AssignTo: &table,
				Columns: []TableViewColumn{
					TableViewColumn{
						Name: "Hash",
					},
					TableViewColumn{
						Name: "Label",
					},
					TableViewColumn{
						Name: "Translation",
					},
				},
				OnItemActivated: tableItemActivated,
			},
			Composite{
				Layout: HBox{
					MarginsZero: true,
					Margins: Margins{
						Left: 6,
					},
				},
				Children: []Widget{
					Label{
						Text: "Search:",
					},
					LineEdit{
						AssignTo: &searchEdit,
						OnTextChanged: func() {
							UpdateShownTableEntries()
						},
					},
				},
			},
		},
		ToolBar: ToolBar{
			ButtonStyle: ToolBarButtonTextOnly,
			Items: []MenuItem{
				Action{
					Text:        "Open",
					OnTriggered: toolOpenTriggered,
					Shortcut:    Shortcut{Modifiers: walk.ModControl, Key: walk.KeyO},
				},
				Action{
					AssignTo:    &saveButton,
					Text:        "Save",
					Enabled:     false,
					OnTriggered: toolSaveTriggered,
					Shortcut:    Shortcut{Modifiers: walk.ModControl, Key: walk.KeyS},
				},
				Action{
					AssignTo:    &addButton,
					Text:        "Add",
					Enabled:     false,
					OnTriggered: toolAddTriggered,
				},
			},
		},
		StatusBarItems: []StatusBarItem{
			StatusBarItem{
				AssignTo: &logStatus,
			},
			StatusBarItem{
				AssignTo: &statusBar,
				Width:    400,
			},
		},
	}.Run()
}
