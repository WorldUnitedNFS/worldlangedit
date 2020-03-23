package main

import (
	"encoding/json"
	"fmt"
	"github.com/WorldUnitedNFS/worldlangedit/lib"
	"github.com/alecthomas/kong"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type LanguagePackJson struct {
	Entries      map[string]string
	SpecialChars []string
}

type Context struct {
	//
}

//noinspection GoStructTag
type UnpackCommand struct {
	InputPath  string `arg name:"in" help:"Path to folder to read language files from."`
	OutputPath string `arg name:"out" help:"Path to folder to generate text files in."`
}

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
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

	matches, err := filepath.Glob(path.Join(r.InputPath, "*_Global.bin"))

	if err != nil {
		panic(err)
	}

	for _, fp := range matches {
		_, fn := filepath.Split(fp)

		if fn == "Largest_Global.bin" || fn == "Labels_Global.bin" {
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
			langJson.Entries[labelsPack.FindEntryByHash(e.Hash).String] = e.String
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

//noinspection GoStructTag
var cli struct {
	Unpack UnpackCommand `cmd help:"Unpack files."`
}

func main() {
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&Context{})
	ctx.FatalIfErrorf(err)
}
