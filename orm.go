package main

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

func NewModel() (*Model, error) {
	return &Model{}, nil
}

type Model struct {
	ID         bson.ObjectId `bson:"_id"`
	Collection string        `bson:"_collection"`
}

func (m Model) Delete() error {
	c := server.Collection(m.Collection)

	if err := c.RemoveId(m.ID); err != nil {
		return errors.New("Could not remove that model")
	}

	return nil
}

func (m *Model) Update(changes bson.M) error {
	c := server.Collection(m.Collection)

	// fmt.Printf("M: %v %T U: %v %T\n", m.ID, m.ID, u.ID, u.ID)
	if err := c.Update(bson.M{"model": bson.M{"_id": m.ID}}, changes); err != nil {
		fmt.Println("Could not update model: ", err)
		// return errors.New("Could not update that model")
		return err
	}

	SetValues(m, changes)

	return nil
}

//x is a double pointer :-)
// pass in User{} and {'_id': 'abcdefghidawdsa', 'Username': ''}
func SetValues(x interface{}, values bson.M) {
	v := reflect.ValueOf(x).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		tag := f.Tag.Get("bson")

		if values[tag] == "" {
			continue
		}

		val := reflect.ValueOf(values[tag])

		v.Field(i).Set(val)
	}
}
