package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	

	"github.com/kartik23-2/Student-API/internal/types"
	"github.com/kartik23-2/Student-API/internal/config"
	_ "modernc.org/sqlite"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config)(*Sqlite, error){
    db, err:=sql.Open("sqlite", cfg.StoragePath)
	if err!=nil{
		return nil, err
	}

	_, err=db.Exec(`CREATE TABLE IF NOT EXISTS students(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	email TEXT,
	age INTEGER
	)`)
	
	if err!=nil{
		return nil, err
	}
	return &Sqlite{
		Db: db,
	}, nil
}

func (s *Sqlite)CreateStudent(name string, email string, age int) (int64, error){
	stmt, err:=s.Db.Prepare("INSERT INTO students (name, email, age) VALUES ($1, $2, $3)")
	if err!=nil{
		return 0, err
	}
	defer stmt.Close()
	
	result, err:=stmt.Exec(name, email, age)
	if err!=nil{
		return 0, err
	}
	lastInsertID, err:=result.LastInsertId()
	if err!=nil{
		return 0, err
	}
	return lastInsertID, nil
}

func (s *Sqlite)GetStudentById(id int64) (types.Student, error){
	stmt, err:=s.Db.Prepare("SELECT name, email, age FROM students WHERE id = $1")
	if err!=nil{
		return types.Student{}, err
	}
	defer stmt.Close()
	
	var student types.Student
	err=stmt.QueryRow(id).Scan(&student.Name, &student.Email, &student.Age)
	if err!=nil{
		if errors.Is(err, sql.ErrNoRows) {
			return types.Student{}, fmt.Errorf("student not found with id:%s", fmt.Sprint(id))
		}
		return types.Student{}, err
	}
	return student, nil
}

func (s *Sqlite)GetAll() ([]types.Student, error){
	stmt, err:=s.Db.Prepare("SELECT name, email, age FROM students")
	if err!=nil{
		return nil, err
	}
	defer stmt.Close()
	
	rows, err:=stmt.Query()
	if err!=nil{
		return nil, err
	}
	defer rows.Close()
	
	var students []types.Student
	for rows.Next() {
		var student types.Student
		err := rows.Scan(&student.Name, &student.Email, &student.Age)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

