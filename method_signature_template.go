package main

import "bytes"

func (s *Struct) SelectAllMethodSignature() string {
	buf := bytes.Buffer{}
	template := newTemplate(`GetAll{{.Name}}() ([]{{.Namespace}}, error)`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) SelectSingleMethodSignature() string {
	buf := bytes.Buffer{}
	template := newTemplate(`Get{{.Name}}({{.PrimaryKeyField.Name}} {{.PrimaryKeyField.Type}}) ({{.Namespace}}, error)`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) InsertMethodSignature() string {
	buf := bytes.Buffer{}
	template := newTemplate(`Insert{{.Name}}({{.ObjectName}} {{.Namespace}}) ({{.Namespace}}, error)`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) UpdateMethodSignature() string {
	buf := bytes.Buffer{}
	template := newTemplate(`Update{{.Name}}({{.ObjectName}} {{.Namespace}}) ({{.Namespace}}, error)`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) DeleteMethodSignature() string {
	buf := bytes.Buffer{}
	template := newTemplate(`Delete{{.Name}}({{.ObjectName}} {{.Namespace}}) ({{.Namespace}}, error)`)
	template.Execute(&buf, s)
	return buf.String()
}
