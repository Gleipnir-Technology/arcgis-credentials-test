package main

import (
	"errors"
	"html/template"
	"io"
	"log"
	"os"
)

var (
	root = newBuiltTemplate("root", "base")
)

type BuiltTemplate struct {
	files []string
	template *template.Template
}

type Link struct {
	Href string
	Title string
}
type ContentRoot struct {
	BabbleLinks []Link
}

func (bt *BuiltTemplate) ExecuteTemplate(w io.Writer, data any) error {
	name := bt.files[0] + ".html"
	if bt.template == nil {
		templ := parseFromDisk(bt.files)
		if templ == nil {
			w.Write([]byte("Failed to read from disk"))
			return errors.New("Template parsing failed")
		}
		return templ.ExecuteTemplate(w, name, data)
	} else {
		return bt.template.ExecuteTemplate(w, name, data)
	}
}

func htmlRoot(w io.Writer, path string) error {
	rootData := ContentRoot{
		BabbleLinks: babbleLinks(path),
	}
	return root.ExecuteTemplate(w, rootData)
}

func makeFuncMap() template.FuncMap {
	funcMap := template.FuncMap{}
	return funcMap
}
func newBuiltTemplate(files ...string) BuiltTemplate {
	files_on_disk := true
	for _, f := range files {
		full_path := "templates/" + f + ".html"
		_, err := os.Stat(full_path)
		if err != nil {
			files_on_disk = false
			break
		}
	}
	if files_on_disk {
		return BuiltTemplate{
			files: files,
			template: nil,
		}
	}
	return BuiltTemplate{
		files: files,
		template: parseEmbedded(files),
	}
}

func parseEmbedded(files []string) *template.Template {
	return nil
}

func parseFromDisk(files []string) *template.Template {
	funcMap := makeFuncMap()
	paths := make([]string, 0)
	for _, f := range files {
		paths = append(paths, "templates/" + f + ".html")
	}
	name := files[0] + ".html"
	templ, err := template.New(name).Funcs(funcMap).ParseFiles(paths...)
	if err != nil {
		log.Println("TEMPLATE FAILED", err)
		return nil
	}
	return templ
}
