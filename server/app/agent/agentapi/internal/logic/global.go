package logic

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

type class map[string]string

var TypeName = class{}
