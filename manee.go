package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	tagRegex       = regexp.MustCompile(`db:"([0-9a-zA-Z,_=&\(\)]*)"`)
	tableNameRegex = regexp.MustCompile(`table:"([0-9a-zA-Z_=&\(\)]*)"`)
)

type PackageFile struct {
	Name        string
	Structs     []Struct
	CreatedDate time.Time
}

type Struct struct {
	CreatedDate time.Time
	Name        string
	Namespace   string
	TableName   string
	Fields      *[]Field
	PackageName string
	ImportPath  string
}

type Field struct {
	Name       string
	Nullable   bool
	Type       string
	ColumnName string
	PrimaryKey bool
	Index      int
}

func (s *Struct) NameLowercase() string {
	return strings.ToLower(s.Name)
}

func (s *Struct) ObjectName() string {
	return strings.ToLower(s.Name)[:1]
}

func (s Struct) CommaSeparatedColumnNamesWithIndex() string {
	cols := []string{}
	for _, f := range *s.Fields {
		cols = append(cols, fmt.Sprintf("%v=$%d", f.ColumnName, f.Index))
	}
	return strings.Join(cols, ", ")
}

func (s Struct) CommaSeparatedColumnNamesWithIndexWithOutPrimaryKey() string {
	cols := []string{}
	for _, f := range *s.Fields {
		if f.PrimaryKey {
			continue
		}
		cols = append(cols, fmt.Sprintf("%v=$%d", f.ColumnName, f.Index))
	}
	return strings.Join(cols, ", ")
}

func (s Struct) CommaSeparatedColumnIndexs() string {
	cols := []string{}
	for _, f := range *s.Fields {
		cols = append(cols, fmt.Sprintf("$%d", f.Index))
	}
	return strings.Join(cols, ", ")
}

func (s Struct) CommaSeparatedColumns() string {
	cols := []string{}
	for _, f := range *s.Fields {
		cols = append(cols, f.ColumnName)
	}
	return strings.Join(cols, ", ")
}

func (s Struct) CommaSeparatedScans() string {
	objs := []string{}
	for _, f := range *s.Fields {
		obj := fmt.Sprintf("&%v.%v", s.ObjectName(), f.Name)
		objs = append(objs, obj)
	}
	return strings.Join(objs, ", ")
}

func (s Struct) CommaSeparatedQueryRow() string {
	objs := []string{}
	for _, f := range *s.Fields {
		obj := fmt.Sprintf("%v.%v", s.ObjectName(), f.Name)
		objs = append(objs, obj)
	}
	return strings.Join(objs, ", ")
}

func (s Struct) CommaSeparatedQueryRowPrimaryKey() string {
	objs := []string{}
	for _, f := range *s.Fields {
		if f.PrimaryKey == false {
			continue
		}
		obj := fmt.Sprintf("%v.%v", s.ObjectName(), f.Name)
		objs = append(objs, obj)
	}
	return strings.Join(objs, ", ")
}

func (s Struct) Returning() string {
	return fmt.Sprintf("returning %v;", s.CommaSeparatedColumns())
}

func (s Struct) PrimaryKeyField() *Field {
	for _, f := range *s.Fields {
		if f.PrimaryKey {
			return &f
		}
	}
	return nil
}

func (s Struct) WhereAtPrimaryKeyLastIndex() string {
	f := s.PrimaryKeyField()
	if f == nil {
		return ""
	}
	return fmt.Sprintf("where %v=$%d", f.ColumnName, len(*s.Fields)+1)
}

func (s Struct) WhereAtPrimaryKeyFirstIndex() string {
	f := s.PrimaryKeyField()
	if f == nil {
		return ""
	}
	return fmt.Sprintf("where %v=$1", f.ColumnName)
}

func (s Struct) WhereAtPrimaryKeyIndex() string {
	f := s.PrimaryKeyField()
	if f == nil {
		return ""
	}
	return fmt.Sprintf("where %v=$%d", f.ColumnName, f.Index)
}

func newTemplate(templateString string) *template.Template {
	return template.Must(template.New("").Parse(templateString))
}

func (p *PackageFile) Write(dir string) {
	baseDir := filepath.Base(filepath.Dir(dir))
	importPath := filepath.Join(baseDir, filepath.Base(dir))
	for _, s := range p.Structs {
		s.ImportPath = importPath
		packageDir := filepath.Join(dir, strings.ToLower(s.Name))

		//create directory
		os.MkdirAll(packageDir, 0777)

		repoFileName := filepath.Join(packageDir, fmt.Sprintf("%v_repository_manee.go", strings.ToLower(s.Name)))
		mockRepoFileName := filepath.Join(packageDir, fmt.Sprintf("%v_mock_manee.go", strings.ToLower(s.Name)))
		serviceFileName := filepath.Join(packageDir, fmt.Sprintf("%v_service_manee.go", strings.ToLower(s.Name)))
		os.Create(repoFileName)
		os.Create(mockRepoFileName)
		os.Create(serviceFileName)

		err := ioutil.WriteFile(repoFileName, []byte(s.GenerateRepository()), 0644)
		if err != nil {
			panic(err)
		} else {
			log.Printf("Write %v", repoFileName)
		}

		err = ioutil.WriteFile(mockRepoFileName, []byte(s.GenerateMockRepository()), 0644)
		if err != nil {
			panic(err)
		} else {
			log.Printf("Write %v", mockRepoFileName)
		}

		err = ioutil.WriteFile(serviceFileName, []byte(s.GenerateServiceManager()), 0644)
		if err != nil {
			panic(err)
		} else {
			log.Printf("Write %v", serviceFileName)
		}
	}
}

type Attributes []string

func (a Attributes) FieldByName(field string) bool {
	for _, v := range a {
		if strings.TrimSpace(v) == strings.TrimSpace(field) {
			return true
		}
	}
	return false
}

func Read(filePath string) (*PackageFile, error) {
	fileName := filePath

	file, err := parser.ParseFile(
		token.NewFileSet(),
		fileName,
		nil,
		parser.ParseComments,
	)
	if err != nil {
		fmt.Errorf("Error parsing '%s': %s", filePath, err)
		return nil, err
	}

	p := PackageFile{}
	p.Name = file.Name.Name
	p.Structs = []Struct{}
	for _, decl := range file.Decls {
		typeDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range typeDecl.Specs {

			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structDecl, isStruct := typeSpec.Type.(*ast.StructType)
			//Check if it's struct
			if !isStruct {
				continue
			}

			s := Struct{}

			s.Name = typeSpec.Name.Name
			s.Namespace = fmt.Sprintf("%v.%v", p.Name, s.Name)
			s.TableName = s.Name
			s.PackageName = p.Name
			//if there is table name definition in comment we then find and assign it here.
			if typeDecl.Doc != nil {
				for _, comment := range typeDecl.Doc.List {
					match := tableNameRegex.FindStringSubmatch(comment.Text)
					if len(match) == 2 {
						s.TableName = match[1]
					}
				}
			}

			s.Fields = &[]Field{}

			fields := structDecl.Fields.List
			for index, field := range fields {
				if field.Tag == nil {
					continue
				}
				match := tagRegex.FindStringSubmatch(field.Tag.Value)
				if len(match) == 2 {
					tags := match[1]
					pointerType, isPointerType := field.Type.(*ast.StarExpr)

					f := Field{}
					f.Name = field.Names[0].Name
					f.Nullable = isPointerType
					f.Index = index + 1
					attributes := strings.Split(tags, ",")
					isPrimary := Attributes(attributes).FieldByName("primary")
					f.PrimaryKey = isPrimary
					if len(attributes) == 1 {
						f.ColumnName = tags
					} else {
						//first in tag is column name
						f.ColumnName = attributes[0]
					}

					if isPointerType {
						f.Type = fmt.Sprintf("%v", pointerType.X)
					} else {
						f.Type = fmt.Sprintf("%v", field.Type)
					}
					*s.Fields = append(*s.Fields, f)
				}
			}
			p.Structs = append(p.Structs, s)
		}
	}
	return &p, nil
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	fileArg := flag.String("f", "", "File path to Go file that contains struct")
	urlPathArg := flag.String("u", "", "URL to Go file that contains struct")
	flag.Parse()

	fileName := filepath.Join(dir, *fileArg)

	if len(*urlPathArg) > 0 {
		resp, err := http.Get(*urlPathArg)
		if err != nil {
			panic(err)
			return
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		tempFile, err := ioutil.TempFile("", "manee")
		ioutil.WriteFile(tempFile.Name(), data, 0644)
		fileName = tempFile.Name()
	}

	p, err := Read(fileName)
	if err != nil {
		panic(err)
	}
	writeToDir := filepath.Dir(fileName)
	p.Write(writeToDir)
}
