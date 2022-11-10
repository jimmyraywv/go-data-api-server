package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type department struct {
	ID        string `json:"id" validate:"required"`
	Name      string `json:"name" validate:"required"`
	ManagerID string `json:"mgrId" validate:"required"`
}

type address struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	County  string `json:"county" validate:"required"`
	State   string `json:"state" validate:"required"`
	Zipcode string `json:"zipcode" validate:"required"`
}

type employee struct {
	ID       string     `json:"id" validate:"required"`
	FName    string     `json:"fname" validate:"required"`
	LName    string     `json:"lname" validate:"required"`
	Sex      string     `json:"sex" validate:"required"`
	DOB      time.Time  `json:"dob" validate:"required"`
	HireDate time.Time  `json:"hireDate" validate:"required"`
	Position string     `json:"position" validate:"required"`
	Salary   uint64     `json:"salary" validate:"required"`
	Dept     department `json:"dept" validate:"required"`
	Address  address    `json:"address" validate:"required"`
}

func (e employee) Json() string {
	out, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	return string(out)
}

type employees map[string]employee

func (e employees) Json() string {
	out, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func main() {
	// read data from file

	csvFile, err := os.Open("./data.data")

	if err != nil {
		fmt.Println(err)
	}

	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	reader.Comma = '|' // Use tab-delimited instead of comma <---- here!

	reader.FieldsPerRecord = -1

	csvData, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	employeeData := make(map[string]employee)

	for _, each := range csvData {
		e := employee{}

		e.ID = each[0]
		e.HireDate, _ = time.Parse("2006-01-02", each[1])
		e.FName = each[2]
		e.Sex = each[3]
		e.LName = each[4]
		e.Position = each[5]
		e.Salary, _ = strconv.ParseUint(each[6], 10, 16)
		e.Dept = department{}
		e.Dept.ID = each[7]
		e.Dept.Name = each[8]
		e.Dept.ManagerID = each[9]
		e.DOB, _ = time.Parse("2006-01-02", each[10])
		e.Address = address{}
		e.Address.Street = each[11]
		e.Address.City = each[12]
		e.Address.County = each[13]
		e.Address.State = each[14]
		e.Address.Zipcode = each[15]

		employeeData[e.ID] = e
	}

	jsondata, err := json.Marshal(employeeData) // convert to JSON

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// sanity check
	// NOTE : You can stream the JSON data to http service as well instead of saving to file
	fmt.Println(string(jsondata))

	// now write to JSON file

	// jsonFile, err := os.Create("./data.json")

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// var oneRecord Employee
	// var allRecords []Employee

	// for _, each := range csvData {
	// 	oneRecord.Name = each[0]
	// 	oneRecord.Age, _ = strconv.Atoi(each[1]) // need to cast integer to string
	// 	oneRecord.Job = each[2]
	// 	allRecords = append(allRecords, oneRecord)
	// }

	// jsondata, err := json.Marshal(allRecords) // convert to JSON

	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// // sanity check
	// // NOTE : You can stream the JSON data to http service as well instead of saving to file
	// fmt.Println(string(jsondata))

	// // now write to JSON file

	// jsonFile, err := os.Create("./data.json")

	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer jsonFile.Close()

	// jsonFile.Write(jsondata)
	// jsonFile.Close()
}
