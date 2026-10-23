package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kartik23-2/Student-API/internal/storage"
	"github.com/kartik23-2/Student-API/internal/types"
	"github.com/kartik23-2/Student-API/internal/utils/response"
)

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {



		var student types.Student
		
		err:=json.NewDecoder(r.Body).Decode(&student)
        if errors.Is(err,io.EOF){
			response.WriteJson(w,http.StatusBadRequest,response.GeneralError(fmt.Errorf("Empty Body")))
			return
		}
		if err != nil{
			response.WriteJson(w,http.StatusBadRequest,response.GeneralError(err))
			return
		}

		//Validation request

		if err := validator.New().Struct(student); err != nil{
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(err.(validator.ValidationErrors)))
			return
		}

		lastID,err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)

		if err != nil{
			slog.Error("creating a student")
			response.WriteJson(w,http.StatusBadRequest,response.GeneralError(err))
			return
		}

		

		slog.Info("student created successfully", slog.Int64("id", lastID))
		response.WriteJson(w,http.StatusOK, map[string]int64{"id": lastID})
	}
}

func GetById(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id:=r.PathValue("id")
		slog.Info("get student by id", slog.String("id", id))
		
	}
}