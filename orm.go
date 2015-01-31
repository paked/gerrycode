package main

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

type Modeler interface {
}

func NewModel() (*Model, error) {
	return &Model{}, nil
}

type Model struct {
	ID         bson.ObjectId
	Collection string
}

func (m Model) Delete() error {
	c := server.Collection(m.Collection)

	if err := c.RemoveId(m.ID); err != nil {
		return errors.New("Could not remove that model")
	}

	return nil
}

func (m Model) Update(changes bson.M) error {
	c := server.Collection(m.Collection)
	if err := c.UpdateId(m.ID, changes); err != nil {
		return errors.New("Could not update that model")
	}

	return nil
}

//x is a double pointer :-)
// pass in User{} and {'ID': 'abcdefghidawdsa', 'Username': ''}
func setValues(x interface{}, values bson.M) {
	v := reflect.ValueOf(x).Elem()
	s := reflect.TypeOf(x)

	for key := range values {
		f := v.FieldByName(key)
		name, _ := s.FieldByName(key).Tag.Get("bson")
	}
}
