package storage

import "github.com/kartik23-2/Student-API/internal/types"

type Storage interface{
	CreateStudent(name string, email string, age int) (int64, error)
	GetStudentById(id int64) (types.Student, error)
} 