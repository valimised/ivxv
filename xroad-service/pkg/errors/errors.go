package errors

import (
	"encoding/json"
	"errors"
)

var (
	ErrNotFound   = errors.New("NOT_FOUND")
	ErrBadRequest = errors.New("BAD_REQUEST")
	ErrVotingEnd  = errors.New("VOTING_END")
)

type FieldError struct {
	Code  string      `json:"code"`
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

func (e FieldError) ToErr() error {
	marshal, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return errors.New(string(marshal))
}
