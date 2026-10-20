package response
import(
	"encoding/json"
	"net/http"
	"github.com/go-playground/validator/v10"
	"fmt"
	"strings"
)

type Response struct{
	Status int `json:"status"`
	Error  string `json:"error"`
}

const (
	StatusOK = "Ok"
	StatusError = "Error"
)


func WriteJson(w http.ResponseWriter, status int, data interface{})error{
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func GeneralError(err error) Response{
	return Response{
		Status: http.StatusBadRequest,
		Error: err.Error(),
	}
}

func ValidationError(errs validator.ValidationErrors) Response{
	var errsMsg []string
	for _,err := range errs{
		switch err.ActualTag() {
			case "required":
				errsMsg = append(errsMsg, fmt.Sprintf("%s is required", err.Field()))
			case "email":
				errsMsg = append(errsMsg, fmt.Sprintf("%s is not valid", err.Field()))
			case "numeric":
				errsMsg = append(errsMsg, fmt.Sprintf("%s is not a number", err.Field()))
			case "min":
				errsMsg = append(errsMsg, fmt.Sprintf("%s is too short", err.Field()))
		}
	}

	return Response{
		Status: http.StatusBadRequest,
		Error: 	strings.Join(errsMsg, ", "),
	}
}