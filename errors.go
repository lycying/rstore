package rstore

import "errors"

var (
	KeyIsNilError   = errors.New("key is nil")
	KeyIsNotInteger = errors.New("value is not an integer or out of range")

	WrongReqArgsNumber   = errors.New("rstore: wrong number of arguments")
	WrongWithScoresSynax = errors.New("hstore : Synax error , should WITHSCORES ?")
	ParseIntError        = errors.New("hstore : Parse int error , check your input")
	ParseFloatError      = errors.New("hstore : Parse float error , check your input")
)
