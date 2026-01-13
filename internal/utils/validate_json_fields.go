package utils

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func BindAndValidateStrict(c *gin.Context, dto interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields() // reject unknown keys

	if err := decoder.Decode(dto); err != nil {
		return err
	}

	// Ensure there’s no extra data after JSON object
	if decoder.More() {
		return errors.New("invalid JSON: multiple objects")
	}

	// Now run validator binding checks (if using `binding:"required"`, etc.)
	if err := binding.Validator.ValidateStruct(dto); err != nil {
		return err
	}
	return nil
}
