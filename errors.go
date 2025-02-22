package sdb

import (
	"fmt"
)

var (
	SelectError = fmt.Errorf("error in getting data from database")
	CreateError = fmt.Errorf("error in inserting data into database")
	UpdateError = fmt.Errorf("error in updating data in database")
	DeleteError = fmt.Errorf("error in deleting data from database")
)
