package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dapetoo/greenlight/internal/validator"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Define an envelop type
type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		//return 0, errors.New(fmt.Sprintf("the %d is invalid", id))
		return 0, fmt.Errorf("the %d is invalid", id)
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//Encode the data to JSON returning error if there was one
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	//
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	//Set a limit on the size of the request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	//Init new json.Decoder and call DisallowUnknownFields() before decoding the request body
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	//Decode the request body into the target destination
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type (at character %q)", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		//if json contains a field that cannot be map to the target destination
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contain unknown key %s", fieldName)

		//Request body exceeds 1MB
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger thanj %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		//Generic error
		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// readString() returns a string value from the query string or the provided default if no matching key could be found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	//Extract the value for a given string from the query string
	s := qs.Get(key)

	//If no key exists/empty, return default value
	if s == "" {
		return defaultValue
	}
	return s
}

// readCSV reads a string value from a string from query string and splits it into a slice on the comma character.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	//Extract the value for a given string from the query string
	csv := qs.Get(key)

	//If no key exists/empty, return default value
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

// readInt() reads a string value from the query string and converts it to an integer before returning. If no matching
// key found, return defaultValue. If the value couldn't be converted to an integer then we record an error message
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	//Extract the value from the string
	s := qs.Get(key)

	//If no key exists/empty, return default value
	if s == "" {
		return defaultValue
	}

	//Convert the value to an int, if fail, add error message to the validator instance and return default value
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	//return the converted integer value
	return i
}

// background helper accepts an arbitrary function as a parameter.
func (app *application) background(fn func()) {
	//Increment the WaitGroup counter
	app.wg.Add(1)
	//Launch a background goroutine
	go func() {
		//Decrement the WaitGroup counter before the goroutine returns
		defer app.wg.Done()
		//Recover any panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		//Execute the arbitrary function that we passed as the parameter
		fn()
	}()
}
