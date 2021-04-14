package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"gitlab.com/sirinibin/go-postgresql-rest/config"
)

// Employee : struct for Employee model
type Employee struct {
	ID        uint64    `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedBy uint64    `bson:"created_by,omitempty" json:"created_by,omitempty"`
	UpdatedBy uint64    `bson:"updated_by,omitempty" json:"updated_by,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Email     string    `bson:"email" json:"email"`
	CreatedAt time.Time `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at,omitempty"`
}

type SearchCriterias struct {
	Page     uint32                 `bson:"page,omitempty" json:"page,omitempty"`
	Size     uint32                 `bson:"size,omitempty" json:"size,omitempty"`
	SearchBy map[string]interface{} `bson:"search_by,omitempty" json:"search_by,omitempty"`
	SortBy   string                 `bson:"sort_by,omitempty" json:"sort_by,omitempty"`
}

var SortFields = map[string]bool{"id": true, "name": true, "email": true, "created_by": true, "updated_by": true}

func ParseString(str string) string {
	singleSpacePattern := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(singleSpacePattern.ReplaceAllString(str, " "))
}
func ValidateSortString(str string) error {

	split := strings.Split(str, ",")
	for _, v := range split {
		split2 := strings.Split(v, " ")
		if len(split2) > 2 {
			return errors.New("Invalid sort value")
		} else if len(split2) == 2 {
			if split2[1] != "asc" && split2[1] != "desc" {
				return errors.New("Invalid sort order value ( Note:use asc or desc, default:asc )")
			}
		}
		fieldName := split2[0]
		if !SortFields[fieldName] {
			return errors.New("Invalid field " + fieldName)
		}

	}
	return nil
}

func FindEmployees(criterias SearchCriterias) (*[]Employee, error) {

	var employees []Employee

	offset := (criterias.Page - 1) * criterias.Size

	searchString := ""

	args := []interface{}{}

	argsIndex := 0

	if len(criterias.SearchBy) > 0 {
		searchString += " WHERE "
		i := 1
		for field, v := range criterias.SearchBy {

			value, ok := v.(string)
			if ok {
				argsIndex++
				if _, err := strconv.Atoi(value); err == nil {
					//Integer
					searchString += field + " = $" + strconv.Itoa(argsIndex) + " "
					args = append(args, value)
				} else {
					//string
					searchString += field + " like $" + strconv.Itoa(argsIndex) + " "
					args = append(args, value+"%")
				}

				if i < len(criterias.SearchBy) {
					searchString += " AND "
				}

			}

			i++
		}
	}

	sortString := ""
	if !govalidator.IsNull(criterias.SortBy) {
		//argsIndex++
		//sortString = " order by  $" + strconv.Itoa(argsIndex) + " "
		//args = append(args, criterias.SortBy)
		sortString = fmt.Sprintf("ORDER BY %s", criterias.SortBy)
	}
	args = append(args, offset)
	args = append(args, criterias.Size)

	query := "SELECT id,name,email,created_by,updated_by,created_at,updated_at FROM employee " + searchString + sortString + " offset $" + strconv.Itoa((argsIndex + 1)) + " limit $" + strconv.Itoa((argsIndex + 2)) + " "
	res, err := config.DB.Query(query, args...)
	defer res.Close()

	if err != nil {
		return nil, err
	}

	for res.Next() {
		var employee Employee

		var createdAt string
		var updatedAt string
		err := res.Scan(&employee.ID, &employee.Name, &employee.Email, &employee.CreatedBy, &employee.UpdatedBy, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		layout := "2006-01-02T15:04:05Z"

		employee.CreatedAt, err = time.Parse(layout, createdAt)
		if err != nil {
			return nil, err
		}

		employee.UpdatedAt, err = time.Parse(layout, updatedAt)
		if err != nil {
			return nil, err
		}

		employees = append(employees, employee)

	}
	return &employees, nil

}

func DeleteEmployee(employeeID uint64) (int64, error) {

	res, err := config.DB.Exec("DELETE from employee where id=$1", employeeID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()

}

func IsEmployeeExists(employeeID uint64) (exists bool, err error) {

	var id uint64

	err = config.DB.QueryRow("SELECT id from employee where id=$1", employeeID).Scan(&id)

	return id != 0, err
}

func FindEmployeeByID(id uint64) (*Employee, error) {

	var createdAt string
	var updatedAt string
	var employee Employee

	err := config.DB.QueryRow("SELECT id,created_by,updated_by,name,email,created_at,updated_at from employee where id=$1", id).Scan(&employee.ID, &employee.CreatedBy, &employee.UpdatedBy, &employee.Name, &employee.Email, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	layout := "2006-01-02T15:04:05Z"

	employee.CreatedAt, err = time.Parse(layout, createdAt)
	if err != nil {
		return &employee, err
	}

	employee.UpdatedAt, err = time.Parse(layout, updatedAt)
	if err != nil {
		return &employee, err
	}

	return &employee, err
}

func (employee *Employee) IsEmailExists() (exists bool, err error) {

	var id uint64

	if employee.ID != 0 {
		//Old Record
		err = config.DB.QueryRow("SELECT id from employee where email=$1 and id!=$2", employee.Email, employee.ID).Scan(&id)
	} else {
		//New Record
		err = config.DB.QueryRow("SELECT id from employee where email=$1", employee.Email).Scan(&id)
	}
	return id != 0, err
}

func (employee *Employee) Validate(w http.ResponseWriter, r *http.Request, scenario string) (errs map[string]string) {

	errs = make(map[string]string)

	if scenario == "update" {
		if employee.ID == 0 {
			errs["id"] = "ID is required"
			return errs
		}
		exists, err := IsEmployeeExists(employee.ID)
		if err != nil || !exists {
			errs["id"] = err.Error()
			return errs
		}

	}

	if govalidator.IsNull(employee.Name) {
		errs["name"] = "Name is required"
	}

	if govalidator.IsNull(employee.Email) {

		errs["username"] = "E-mail is required"
	}

	emailExists, err := employee.IsEmailExists()
	if err != nil && err != sql.ErrNoRows {
		errs["email"] = err.Error()
	}

	if emailExists {
		errs["email"] = "E-mail is Already in use"
	}

	if emailExists {
		w.WriteHeader(http.StatusConflict)
	} else if len(errs) > 0 {
		w.WriteHeader(http.StatusBadRequest)
	}

	return errs
}

func (employee *Employee) Insert() error {

	lastInsertId := 0
	err := config.DB.QueryRow("insert into employee (name,created_by,updated_by, email,created_at,updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", employee.Name, employee.CreatedBy, employee.UpdatedBy, employee.Email, employee.CreatedAt, employee.UpdatedAt).Scan(&lastInsertId)
	if err != nil {
		return err
	}

	employee.ID = uint64(lastInsertId)
	log.Print("user.ID:")
	log.Print(employee.ID)

	return nil
}

func (employee *Employee) Update() (*Employee, error) {

	res, err := config.DB.Exec("UPDATE employee SET name=$1, updated_by=$2 ,email=$3, updated_at=$4 WHERE id=$5", employee.Name, employee.UpdatedBy, employee.Email, employee.UpdatedAt, employee.ID)
	if err != nil {
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when finding rows affected", err)
		return nil, err
	}

	employee, err = FindEmployeeByID(employee.ID)
	if err != nil {
		return nil, err
	}

	log.Print("user.ID:")
	log.Print(employee.ID)
	log.Printf("%d employee updated ", rows)

	return employee, nil
}
