package main

import (
	"encoding/json"
	"fmt"
	"github.com/WorldUnitedNFS/worldlangedit/lib"
	"github.com/WorldUnitedNFS/worldlangedit/lib/charmap"
	"github.com/alecthomas/kong"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func BuildCharMap(chars []rune) *charmap.Charmap {
	newCharMap := &charmap.Charmap{
		NumEntries: 0,
		EntryTable: [3072]uint16{},
	}

	numEntries := int32(0x80) // 0x00-0x7F get reserved spaces
	numEntries += int32(len(chars))

	tmpNumEntries := numEntries

	jumpEntries := make([]uint16, 0)
	maxJumpEntry := tmpNumEntries >> 7

	if maxJumpEntry >= 2 {
		jumpEntries = append(jumpEntries, uint16(maxJumpEntry))
		tmpNumEntries++
	}

	// Determine jump entries
	for {
		newMaxJumpEntry := tmpNumEntries >> 7

		if newMaxJumpEntry > maxJumpEntry {
			tmpNumEntries++
			maxJumpEntry = newMaxJumpEntry
			jumpEntries = append(jumpEntries, uint16(maxJumpEntry))
		} else {
			break
		}
	}

	mapIndex := 0x80

	numEntries += maxJumpEntry - 1

	for i := maxJumpEntry; i >= 2; i-- {
		newCharMap.EntryTable[mapIndex] = uint16(i)
		mapIndex++
	}

	for _, r := range chars {
		newCharMap.EntryTable[mapIndex] = uint16(r)
		mapIndex++
	}

	newCharMap.NumEntries = numEntries

	return newCharMap
}

type LanguagePackJson struct {
	Entries      map[string]string
	SpecialChars []string
}

type Context struct {
	//
}

//noinspection GoStructTag
type UnpackCommand struct {
	InputPath  string `arg name:"in" help:"Path to folder to read binary files from."`
	OutputPath string `arg name:"out" help:"Path to folder to generate text files in."`
}

//noinspection GoStructTag
type PackCommand struct {
	InputPath  string `arg name:"in" help:"Path to folder to read text files from."`
	OutputPath string `arg name:"out" help:"Path to folder to generate binary files in."`
}

func (r *UnpackCommand) Run(_ *Context) error {
	if _, err := os.Stat(r.OutputPath); os.IsNotExist(err) {
		_ = os.Mkdir(r.OutputPath, 0644)
	}

	labelsData, err := ioutil.ReadFile(path.Join(r.InputPath, "Labels_Global.bin"))

	if err != nil {
		panic(err)
	}

	labelsPack := lib.ParseFile(labelsData)
	labelMap := make(map[uint32]string)

	for _, e := range labelsPack.Entries {
		labelMap[e.Hash] = e.String
	}

	matches, err := filepath.Glob(path.Join(r.InputPath, "*_Global.bin"))

	if err != nil {
		panic(err)
	}

	for _, fp := range matches {
		_, fn := filepath.Split(fp)

		if fn == "Largest_Global.bin" {
			continue
		}

		fmt.Println("Loading file:", fp)
		enc, err := ioutil.ReadFile(fp)
		if err != nil {
			panic(err)
		}
		langFile := lib.ParseFile(enc)
		fmt.Printf("Loaded %d strings from file\n", len(langFile.Entries))
		cleanName := FilenameWithoutExtension(fn)
		jsonName := path.Join(r.OutputPath, cleanName+".json")
		langJson := LanguagePackJson{
			Entries:      make(map[string]string),
			SpecialChars: make([]string, 0),
		}

		for _, e := range langFile.Entries {
			langJson.Entries[labelMap[e.Hash]] = e.String
		}

		for _, c := range langFile.CharMap.EntryTable {
			if c >= 0x80 {
				langJson.SpecialChars = append(langJson.SpecialChars, string(rune(c)))
			}
		}

		f, err := os.Create(jsonName)
		if err != nil {
			panic(err)
		}
		encoder := json.NewEncoder(f)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", " ")
		err = encoder.Encode(langJson)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (r *PackCommand) Run(_ *Context) error {
	if _, err := os.Stat(r.OutputPath); os.IsNotExist(err) {
		_ = os.Mkdir(r.OutputPath, 0644)
	}

	matches, err := filepath.Glob(path.Join(r.InputPath, "*_Global.json"))

	if err != nil {
		panic(err)
	}

	labelJson := LoadLanguageJson(path.Join(r.InputPath, "Labels_Global.json"))
	labelPack := BuildLangFileFromJson(labelJson)

	type SavePackEntry struct {
		Name string
		Pack *lib.LangFile
	}

	packs := []SavePackEntry{{
		Name: "Labels_Global",
		Pack: labelPack,
	}}

	for _, fp := range matches {
		_, fn := filepath.Split(fp)

		if strings.Contains(fn, "Labels") {
			continue
		}
		cleanName := FilenameWithoutExtension(fn)

		langJson := LoadLanguageJson(fp)

		fmt.Println("Loaded", len(langJson.Entries), "strings from", fp)

		lp := BuildLangFileFromJson(langJson)
		packs = append(packs, SavePackEntry{
			Name: cleanName,
			Pack: lp,
		})
	}

	for _, p := range packs {
		op := path.Join(r.OutputPath, p.Name+".bin")
		fmt.Println("Saving", p.Name, "to", op)
		f := lib.SaveFile(p.Pack, labelPack, true)
		err = ioutil.WriteFile(op, f, 666)
		if err != nil {
			panic(err)
		}
		fmt.Println("Saved", p.Name, "to", op)
	}

	return nil
}

func BuildLangFileFromJson(langJson *LanguagePackJson) *lib.LangFile {
	specialChars := make([]rune, 0)

	for _, e := range langJson.SpecialChars {
		var first rune
		for _, c := range e {
			first = c
			break
		}

		specialChars = append(specialChars, first)
	}

	entries := make([]lib.LangFileEntry, 0)

	for l, e := range langJson.Entries {
		entries = append(entries, lib.LangFileEntry{
			Hash:          lib.BinHash(l),
			String:        e,
			Offset:        0,
			OriginalBytes: nil,
		})
	}

	lp := &lib.LangFile{
		Entries: entries,
		CharMap: BuildCharMap(specialChars),
	}
	return lp
}

func LoadLanguageJson(fp string) *LanguagePackJson {
	f, err := os.Open(fp)

	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	langJson := &LanguagePackJson{}
	err = decoder.Decode(&langJson)

	if err != nil {
		panic(err)
	}

	return langJson
}

//noinspection GoStructTag
var cli struct {
	Unpack UnpackCommand `cmd help:"Unpack files."`
	Pack   PackCommand   `cmd help:"Pack files."`
}

func main() {
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&Context{})
	ctx.FatalIfErrorf(err)
}