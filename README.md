# Manee
This little Go program helps generating CRUD operations from your Go's struct.  
Manee uses Go standard package: AST to read your struct code and struct field tags.  
**I use PostgreSQL so this was tested only on PostgreSQL**
**Go struct file must be valid and compilable**


## Usage
Use **// table:"table_name"** to define your table name.  
    If not specifiy, Manee assumes that your table has the same name as your struct name.  
Use **db:"field_name"** to define your field name.  
    Manee will skip struct property which doesn't have **db** tag in it.  
Use **primary** tag in struct tag to specify primary key.  

## Sample struct
```go
package models

import "time"

//  table:"table_person"
//  another line of comment passing by.
type Person struct {
    ID        string    `json:"id,omitempty" db:"id,primary"`
    Name      *string   `json:"name,omitempty" db:"name"`
    Address   string    `json:"address,omitempty" db:"address"`
    Published *bool     `json:"published,omitempty" db:"published"`
    Created   time.Time `json:"created,omitempty" db:"created"`
}
```

## Run
./manee -f=yourStructFile.go

## Result
Manee will create a new directory where your struct file is located.  
Directory contains 3 files.  
*struct_name*_mock_manee.go  
*struct_name*_repository_manee.go  
*struct_name*_service_manee.go  

Manee uses repository and dependency injection pattern.  

Open *struct_name*_service_manee.go and you will see  

```go
    //m := &PersonRepository{DB: nil}
    m := &MockPersonRepository{}
```  
When you want to test with your database, simply uncomment and inject your *sql.DB connection into it.




