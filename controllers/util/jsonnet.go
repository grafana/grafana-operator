package util

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/go-jsonnet"
	"github.com/grafana-operator/grafana-operator/v5/embeds"
)

// EmbedFSImporter "imports" data from an in-memory embedFS.
type EmbedFSImporter struct {
	Embed embed.FS
	cache map[string]jsonnet.Contents
	mutex sync.Mutex
}

// Import fetches data from an embedFS struct.
func (importer *EmbedFSImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	importer.mutex.Lock()
	defer importer.mutex.Unlock()

	if importer.cache == nil {
		importer.cache = make(map[string]jsonnet.Contents)
	}

	fetchContents := func(getPath, foundAt string) (contents jsonnet.Contents, found string, err error) {
		if content, ok := importer.cache[getPath]; ok {
			return content, getPath, nil
		}

		b, err := importer.Embed.ReadFile(getPath)
		if err != nil {
			return jsonnet.Contents{}, "", err
		}

		file, err := importer.Embed.Open(getPath)
		if err != nil {
			return jsonnet.Contents{}, "", err
		}
		defer func(file fs.File) {
			_ = file.Close()
		}(file)

		importer.cache[foundAt] = jsonnet.MakeContentsRaw(b)

		return jsonnet.MakeContentsRaw(b), foundAt, nil
	}

	var foundContents jsonnet.Contents
	var s string

	findImport := func(root string) error {
		err = fs.WalkDir(importer.Embed, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, importedPath) {
				foundContents, s, err = fetchContents(path, importedPath)
				if err != nil {
					return err
				}
				return filepath.SkipDir
			}
			return nil
		})
		return err
	}

	err = findImport(filepath.Join("grafonnet-lib", "grafonnet"))
	if err != nil || foundContents.Data() != nil {
		return foundContents, s, err
	}

	err = findImport(filepath.Join("grafonnet-lib", "grafonnet-7.0"))
	if err != nil {
		return jsonnet.Contents{}, "", err
	}

	return foundContents, s, nil
}

func FetchJsonnet(snippet string) ([]byte, error) {
	vm := jsonnet.MakeVM()

	vm.Importer(&EmbedFSImporter{Embed: embeds.GrafonnetEmbed})

	jsonString, err := vm.EvaluateAnonymousSnippet("", snippet)
	return []byte(jsonString), err
}
