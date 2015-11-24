package main

import "bytes"

func (s *Struct) GenerateServiceManager() string {
	buf := bytes.Buffer{}

	template := newTemplate(`// This file is generated by Manee.
// {{.CreatedDate}}
package {{.NameLowercase}}

import "{{.ImportPath}}"

type {{.Name}}Service struct {
	{{.Name}}ServiceManager
}

type {{.Name}}ServiceManager interface {
	{{.SelectAllMethodSignature}}
	{{.SelectSingleMethodSignature}}
	{{.InsertMethodSignature}}
	{{.UpdateMethodSignature}}
	{{.DeleteMethodSignature}}
}

func New{{.Name}}Service() *{{.Name}}Service {
	//m := &{{.Name}}Repository{DB: nil}
	m := &Mock{{.Name}}Repository{}
	return &{{.Name}}Service{ {{.Name}}ServiceManager: m }
}`)
	template.Execute(&buf, s)
	return buf.String()
}
