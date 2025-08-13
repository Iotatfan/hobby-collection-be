package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/iotatfan/hobby-collection-be/internal/text"
	"gorm.io/gorm"
)

func ParseError(errs ...error) (string, int) {
	badReqCode := http.StatusBadRequest
	for _, err := range errs {
		switch typedError := any(err).(type) {
		case validator.ValidationErrors:
			return parseFieldError(typedError[0]), badReqCode
		case *json.UnmarshalTypeError:
			return parseMarshallingError(*typedError), badReqCode
		case ValError:
			return typedError.Error(), badReqCode
		case *BcryptError:
			log.Printf("Bcrypt Error: %s\n", typedError.ErrorMsg.Error())
			return text.ErrServer, http.StatusInternalServerError
		case *DBError:
			return parseDBError(*typedError)
		case NoRouteError:
			return typedError.ErrorMsg, http.StatusNotFound
		case JWTError:
			return typedError.ErrorMsg, http.StatusUnauthorized
		case ServiceError:
			return typedError.ErrorMsg, typedError.Code
		case *strconv.NumError:
			return parseStrConvError(typedError), badReqCode
		case *time.ParseError:
			return fmt.Sprintf("%s is not a valid date", typedError.ValueElem), badReqCode
		case nil:
			// do nothing
		default:
			return err.Error(), http.StatusInternalServerError
		}
	}
	return "success", http.StatusOK
}

func parseDBError(e DBError) (string, int) {
	switch e.ErrorMsg {
	case gorm.ErrRecordNotFound:
		return "record not found", http.StatusNotFound
	default:
		log.Printf("%s", e.ErrorMsg.Error())
		return text.ErrServer, http.StatusInternalServerError
	}
}

func parseFieldError(e validator.FieldError) string {
	fieldName := fmt.Sprintf("the field %s", e.Field())
	tag := e.Tag()
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return fmt.Sprintf("%s must be a valid email", fieldName)
	case "alpha":
		return fmt.Sprintf("%s must be alphabet only", fieldName)
	case "oneof":
		return fmt.Sprintf("%s must be one of %s", fieldName, e.Param())
	case "unique":
		return fmt.Sprintf("%s from %s must not have duplicates", e.Param(), fieldName)
	case "min":
		return fmt.Sprintf("%s must be %s or higher", fieldName, e.Param())
	case "max":
		return fmt.Sprintf("%s must be equal or lower than %s characters", fieldName, e.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid url", fieldName)
	default:
		return fmt.Sprintf("%v", e)
	}
}
func parseMarshallingError(e json.UnmarshalTypeError) string {
	return fmt.Sprintf("the field %s must be %s", e.Field, e.Type.String())
}

func parseStrConvError(err *strconv.NumError) string {
	if err.Func == "ParseBool" {
		return fmt.Sprintf("%s is not a boolean", err.Num)
	}
	if err.Func == "ParseInt" {
		return fmt.Sprintf("%s is not an integer", err.Num)
	}
	if err.Func == "ParseFloat" {
		return fmt.Sprintf("%s is not a float", err.Num)
	}
	return fmt.Sprintf("%s is not a number", err.Num)
}
