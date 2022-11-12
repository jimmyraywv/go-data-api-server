package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	Log "github.com/sirupsen/logrus"
	"io"
	"jimmyray.io/data-api/pkg/data"
	"jimmyray.io/data-api/pkg/utils"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/go-playground/validator/v10"
)

const (
	HttpReqReadErr    string = "HTTP_REQ_READ_ERR"
	JsonEncodeErr     string = "JSON_ENCODE_ERR"
	DataConflictErr   string = "NOOP_DATA_CONFLICT_ERR"
	DataNotFoundErr   string = "NOOP_DATA_NOT_FOUND_ERR"
	ValidationErr     string = "VALIDATION_ERR"
	InternalServerErr string = "INTERNAL_SERVER_ERR"
	MockDataErr       string = "Mock_Data_Err"
	IncorrectIdErr    string = "INCORRECT_ID_ERR"
	IncorrectInputErr string = `Please check submission:
{"id":"<id>","fname":"<fname>","lname":"<lanme>","sex":"<sex>","dob":"<yyyy-mm-ddThh:MM:ssZ>",
"hireDate":"<yyyy-mm-ddThh:MM:ssZ>","position":"<position>>","salary":<salary>,
"dept":{"id":"<id>","name":"<name>","mgrId":"<mgrId>"},
"address":{"street":"<street>","city":"<city>","county":"<county>","state":"<st>","zipcode":"<00000>"}}`
)

var (
	ErrDataNotFound = errors.New(DataNotFoundErr)
	ErrDataConflict = errors.New(DataConflictErr)
)

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
}

var serviceId = uuid.New()

func GetServiceId() string {
	return serviceId.String()
}

// Employee
type Department struct {
	ID        string `json:"id" validate:"required"`
	Name      string `json:"name" validate:"required"`
	ManagerID string `json:"mgrId" validate:"required"`
}

type Address struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	County  string `json:"county" validate:"required"`
	State   string `json:"state" validate:"required"`
	Zipcode string `json:"zipcode" validate:"required"`
}

type Employee struct {
	ID       string     `json:"id" validate:"required"`
	FName    string     `json:"fname" validate:"required"`
	LName    string     `json:"lname" validate:"required"`
	Sex      string     `json:"sex" validate:"required"`
	DOB      time.Time  `json:"dob" validate:"required"`
	HireDate time.Time  `json:"hireDate" validate:"required"`
	Position string     `json:"position" validate:"required"`
	Salary   uint64     `json:"salary" validate:"required"`
	Dept     Department `json:"dept" validate:"required"`
	Address  Address    `json:"address" validate:"required"`
}

func (e Employee) Json() string {
	out, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	return string(out)
}

type Employees map[string]Employee

func (e Employees) Json() string {
	out, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	return string(out)
}

type Logic struct {
	ServiceData Employees
	m           sync.Mutex
}

type Controller struct {
	L        *Logic
	Validate *validator.Validate
}

type info struct {
	NAME string `json:"service-name"`
	ID   string `json:"service-id"`
}

func (i info) String() string {
	out, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}

	return string(out)
}

var ServiceInfo = info{}

type InfoController struct {
	ServiceInfo info
}

var (
	L  Logic
	IC InfoController
	C  Controller
)

func (e Employees) search(id string) (Employee, bool) {
	d, found := e[id]
	return d, found
}

func (l *Logic) Create(newData Employee) error {
	l.m.Lock()
	defer l.m.Unlock()

	if _, found := l.ServiceData.search(newData.ID); found {
		return ErrDataConflict
	}

	l.ServiceData[newData.ID] = newData
	return nil
}

func (l *Logic) Read(id string) (Employee, bool) {
	l.m.Lock()
	defer l.m.Unlock()

	return l.ServiceData.search(id)
}

func ReadAll() Employees {
	L.m.Lock()
	defer L.m.Unlock()

	// returning a copy
	out := Employees{}
	for k, v := range L.ServiceData {
		out[k] = v
	}

	return out
}

func (l *Logic) Update(input Employee) (Employee, error) {
	l.m.Lock()
	defer l.m.Unlock()

	foundData, found := l.ServiceData[input.ID]
	if !found {
		return foundData, ErrDataNotFound
	}
	if foundData == input {
		return foundData, ErrDataConflict
	}
	l.ServiceData[input.ID] = input
	return l.ServiceData[input.ID], nil
}

func Delete(id string) error {
	L.m.Lock()
	defer L.m.Unlock()

	if _, found := L.ServiceData[id]; !found {
		return ErrDataNotFound
	}

	delete(L.ServiceData, id)
	return nil
}

func (ic InfoController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "OK")
}

func (ic InfoController) GetServiceInfo(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, ServiceInfo.String())
}

func LoadMockData() error {
	e := Employees{}
	err := json.Unmarshal([]byte(data.MockData), &e)

	if err != nil {
		utils.Logger.WithFields(Log.Fields{"error": err.Error()}).Debug("")
	}

	utils.Logger.WithFields(Log.Fields{"length": len(e)}).Debug("Employee map length")

	if err == nil {
		L.m.Lock()
		defer L.m.Unlock()

		for k, v := range e {
			L.ServiceData[k] = v
		}
	}

	utils.Logger.WithFields(Log.Fields{"length": len(L.ServiceData)}).Debug("ServiceData map length")

	return err
}

func (c Controller) DeleteData(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	id := mux.Vars(r)["id"]
	err := validateId(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, IncorrectIdErr)
		return
	}

	err = Delete(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, ErrDataNotFound)
	}
}

func (c Controller) PatchData(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	var input Employee

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		errorData := utils.ErrorLog{Skip: 1, Event: HttpReqReadErr, Message: err.Error()}
		utils.LogErrors(errorData)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, IncorrectInputErr)
		return
	}

	err = c.Validate.Struct(input)
	if err != nil {
		errorData := utils.ErrorLog{Skip: 1, Event: ValidationErr, Message: err.Error(), ErrorData: string(input.Json())}
		utils.LogErrors(errorData)

		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, IncorrectInputErr)
		return
	}

	err = validateId(input.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, IncorrectIdErr)
		return
	}

	updated, err := c.L.Update(input)
	if err != nil {
		if errors.Is(err, ErrDataConflict) {
			w.WriteHeader(http.StatusConflict)
			_, _ = fmt.Fprint(w, ErrDataConflict)
			return
		}
		if errors.Is(err, ErrDataNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, ErrDataNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(updated)
	if err != nil {
		errorData := utils.ErrorLog{Skip: 1, Event: JsonEncodeErr, Message: err.Error(), ErrorData: string(updated.Json())}
		utils.LogErrors(errorData)
	}
}

func (c Controller) CreateData(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	var newData Employee

	err := json.NewDecoder(r.Body).Decode(&newData)
	if err != nil {
		errorData := utils.ErrorLog{Skip: 1, Event: HttpReqReadErr, Message: err.Error()}
		utils.LogErrors(errorData)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(IncorrectInputErr))
		return
	}

	err = C.Validate.Struct(newData)
	if err != nil {
		errorData := utils.ErrorLog{Skip: 1, Event: HttpReqReadErr, Message: err.Error(), ErrorData: string(newData.Json())}
		utils.LogErrors(errorData)

		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, IncorrectInputErr)
		return
	}

	err = validateId(newData.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, IncorrectIdErr)
		return
	}

	err = C.L.Create(newData)

	if err != nil {
		if errors.Is(err, ErrDataConflict) {
			w.WriteHeader(http.StatusConflict)
			_, _ = fmt.Fprint(w, DataConflictErr)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, InternalServerErr)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (c Controller) GetData(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	err := validateId(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, IncorrectIdErr)
		return
	}

	foundData, found := c.L.Read(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, DataNotFoundErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(foundData)
	if err != nil {
		errorData := utils.ErrorLog{Skip: 1, Event: JsonEncodeErr, Message: err.Error(), ErrorData: string(foundData.Json())}
		utils.LogErrors(errorData)
	}
}

func (c Controller) GetAllData(w http.ResponseWriter, r *http.Request) {
	allData := ReadAll()
	err := json.NewEncoder(w).Encode(allData)
	if err != nil {
		errorData := utils.ErrorLog{Skip: 1, Event: JsonEncodeErr, Message: err.Error(), ErrorData: string(allData.Json())}
		utils.LogErrors(errorData)
	}
}

func validateId(id string) error {
	_, err := strconv.ParseInt(id, 10, 64)
	return err
}
