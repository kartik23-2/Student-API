package student

import (
	"fmt"
	"encoding/json"
	"errors"
	"io"
	"github.com/kartik23-2/Student-API/internal/types"
	"github.com/kartik23-2/Student-API/internal/utils/response"
    "github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

func New() http.HandlerFunc {
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



		slog.Info("creating a student")
		response.WriteJson(w,http.StatusOK,"student created successfully")
	}
}