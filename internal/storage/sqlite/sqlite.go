package sqlite

import (
	"database/sql"

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