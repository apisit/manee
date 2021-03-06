package main

import (
	"bytes"
	"time"
)

func (s *Struct) GenerateRepository() string {
	s.CreatedDate = time.Now()
	buf := bytes.Buffer{}
	template := newTemplate(`// This file is generated by Manee.
// {{.CreatedDate}}
package {{.NameLowercase}}

import (
	"{{.ImportPath}}"
	"database/sql"
)

type {{.Name}}Repository struct {
	DB *sql.DB
}
{{.SelectAllTemplate}}
{{.SelectTemplate}}
{{.InsertTemplate}}
{{.UpdateTemplate}}
{{.DeleteTemplate}}
`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) InsertStatement() string {
	buf := bytes.Buffer{}
	template := newTemplate(`INSERT INTO {{.TableName}}({{.CommaSeparatedColumns}}) VALUES({{.CommaSeparatedColumnIndexs}}) {{.Returning}}`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) UpdateStatement() string {
	buf := bytes.Buffer{}
	template := newTemplate(`UPDATE {{.TableName}} SET {{.CommaSeparatedColumnNamesWithIndexWithOutPrimaryKey}} {{.WhereAtPrimaryKeyIndex}} {{.Returning}}`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) DeleteStatement() string {
	buf := bytes.Buffer{}
	template := newTemplate(`DELETE from {{.TableName}} {{.WhereAtPrimaryKeyFirstIndex}} {{.Returning}}`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) SelectStatement() string {
	buf := bytes.Buffer{}
	template := newTemplate("SELECT {{.CommaSeparatedColumns}} FROM {{.TableName}};")
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) SelectSingleStatement() string {
	buf := bytes.Buffer{}
	template := newTemplate("SELECT {{.CommaSeparatedColumns}} FROM {{.TableName}} {{.WhereAtPrimaryKeyFirstIndex}};")
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) InsertTemplate() string {
	buf := bytes.Buffer{}
	template := newTemplate(`
func (d *{{.Name}}Repository) Insert{{.Name}}({{.ObjectName}} {{.Namespace}}) ({{.Namespace}}, error) {
	stmt, err := d.DB.Prepare(` + "`" + `{{.InsertStatement}}` + "`)" + `
	if err != nil {
		return {{.ObjectName}}, err
	}
	defer stmt.Close()
	err = stmt.QueryRow({{.CommaSeparatedQueryRow}}).Scan({{.CommaSeparatedScans}})
	if err != nil {
		return {{.ObjectName}}, err
	}
	return {{.ObjectName}}, nil
}`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) UpdateTemplate() string {
	buf := bytes.Buffer{}
	template := newTemplate(`
func (d *{{.Name}}Repository) Update{{.Name}}({{.ObjectName}} {{.Namespace}}) ({{.Namespace}}, error) {
	stmt, err := d.DB.Prepare(` + "`" + `{{.UpdateStatement}}` + "`)" + `
	if err != nil {
		return {{.ObjectName}}, err
	}
	defer stmt.Close()
	err = stmt.QueryRow({{.CommaSeparatedQueryRow}}).Scan({{.CommaSeparatedScans}})
	if err != nil {
		return {{.ObjectName}}, err
	}
	return {{.ObjectName}}, nil
}`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) DeleteTemplate() string {
	buf := bytes.Buffer{}
	template := newTemplate(`
func (d *{{.Name}}Repository) Delete{{.Name}}({{.ObjectName}} {{.Namespace}}) ({{.Namespace}}, error) {
	stmt, err := d.DB.Prepare(` + "`" + `{{.DeleteStatement}}` + "`)" + `
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRow({{.CommaSeparatedQueryRowPrimaryKey}}).Scan({{.CommaSeparatedScans}})
	if err != nil {
		return nil, err
	}
	return {{.ObjectName}}, nil
}`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) SelectTemplate() string {
	buf := bytes.Buffer{}
	template := newTemplate(`
func (d *{{.Name}}Repository) Get{{.Name}}({{.PrimaryKeyField.Name}} {{.PrimaryKeyField.Type}}) ({{.Namespace}}, error) {
	{{.ObjectName}} := {{.Namespace}}{}
	stmt, err := d.DB.Prepare(` + "`" + `{{.SelectSingleStatement}}` + "`)" + `
	if err != nil {
		return {{.ObjectName}}, err
	}
	defer stmt.Close()
	err = stmt.QueryRow({{.PrimaryKeyField.Name}}).Scan({{.CommaSeparatedScans}})
	if err != nil {
		return {{.ObjectName}}, err
	}
	return {{.ObjectName}}, nil
}`)
	template.Execute(&buf, s)
	return buf.String()
}

func (s *Struct) SelectAllTemplate() string {
	buf := bytes.Buffer{}
	template := newTemplate(`
func (d *{{.Name}}Repository) GetAll{{.Name}}() ([]{{.Namespace}}, error) {
	list := []{{.Namespace}}{}
	stmt, err := d.DB.Prepare(` + "`" + `{{.SelectStatement}}` + "`)" + `
	if err != nil {
		return list, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		{{.ObjectName}} := {{.Namespace}}{}
		errScan := rows.Scan({{.CommaSeparatedScans}})
		if errScan != nil {
			return list, err
		}
		list = append(list, {{.ObjectName}})
	}
	return list, nil
}`)
	template.Execute(&buf, s)
	return buf.String()
}
