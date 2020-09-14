//+build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	embeddedFileName = "embedded_resources.go"
	staticFolder     = "../frontend/build"
)

func bytesToEmbeddedByteArray(data []byte) string {
	builder := strings.Builder{}
	for _, v := range data {
		builder.WriteString(fmt.Sprintf("%d,", int(v)))
	}
	return builder.String()
}

func sizeOfBytes(data []byte) int {
	return len(data)
}

var conv = map[string]interface{}{"conv": bytesToEmbeddedByteArray, "len": sizeOfBytes}
var tmpl = template.Must(template.New("").Funcs(conv).Parse(`package embed

{{- range $res := . }}
var res{{ $res.Index }} = []byte{ {{ conv $res.Data }} }
{{- end }}

func init() {
	{{- range $res := . }}
		Add("{{ $res.Name }}", res{{ $res.Index }})
	{{- end }}
}

`))

type res struct {
	Name  string
	Data  []byte
	Index int
}

func main() {
	if _, err := os.Stat(staticFolder); os.IsNotExist(err) {
		log.Fatal("Configs directory does not exists!")
	}

	var files []res

	prefix, _ := filepath.Abs(staticFolder)
	err := filepath.Walk(staticFolder, func(path string, info os.FileInfo, err error) error {
		absPath, _ := filepath.Abs(path)
		relPath := filepath.ToSlash(strings.TrimPrefix(absPath, prefix))
		if info.IsDir() {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read %s: %s", path, err)
			return err
		}
		log.Printf("Embedding: %s", relPath)
		files = append(files, res{Name: relPath, Data: b, Index: len(files)})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	builder := &bytes.Buffer{}
	if err = tmpl.Execute(builder, files); err != nil {
		log.Fatalf("Failed to embed files: %s", err)
	}

	gosrc, err := format.Source(builder.Bytes())
	if err != nil {
		log.Fatalf("Failed to format generated code: %s", err)
	}

	if err = ioutil.WriteFile(embeddedFileName, gosrc, os.ModePerm); err != nil {
		log.Fatalf("Error while saving %s: %s", embeddedFileName, err)
	}
}
