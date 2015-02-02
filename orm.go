package main

import (
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

// Modeller is an interface for use with the ORM, describing a model.
type Modeller interface {
	ID() bson.ObjectId
	C() string
}

// UpdateModel updates a Modeller interface with the provided values in persistent storage.
// It is an alias function for UpdateModel, and then UpdateValues.
func UpdateModel(m Modeller, values bson.M) error {
	if err := UpdateValues(m, values); err != nil {
		return err
	}

	SetValues(m, values)

	return nil
}

// UpdateValues updates a model in the MongoDB.
func UpdateValues(m Modeller, values bson.M) error {
	c := server.Collection(m.C())

	return c.UpdateId(m.ID(), bson.M{"$set": values})
}

// Restore a model from a persisted MongoDB record.
func RestoreModel(m Modeller, id bson.ObjectId) error {
	c := server.Collection(m.C())

	return c.FindId(id).One(m)
}

// pass in User{} and {'_id': 'abcdefghidawdsa', 'Username': ''}
func SetValues(x interface{}, values bson.M) {
	v := reflect.ValueOf(x).Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		tag := f.Tag.Get("bson")

		val := reflect.ValueOf(values[tag])

		if !val.IsValid() || empty(val) {
			continue
		}

		v.Field(i).Set(val)
	}
}

func empty(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
