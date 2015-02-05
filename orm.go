package main

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

// Modeller is an interface for use with the ORM, describing a model.
type Modeller interface {
	BID() bson.ObjectId
	C() string
}

//PersistModel creates a copy of the model and persists it in the DB.
func PersistModel(m Modeller) error {
	c := server.Collection(m.C())

	if err := c.Insert(m); err != nil {
		return err
	}

	return nil
}

// UpdateModel updates a Modeller interface with the provided values in persistent storage.
// It is an alias function for UpdateModel, and then UpdateValues.
func UpdateModel(m Modeller, values bson.M) error {
	if err := updateValues(m, values); err != nil {
		return err
	}

	setValues(m, values)

	return nil
}

// Remove removes a model from the MongoDB.
func RemoveModel(m Modeller) error {
	c := server.Collection(m.C())

	return c.RemoveId(m.BID())
}

// UpdateValues updates a model in the MongoDB.
func updateValues(m Modeller, values bson.M) error {
	c := server.Collection(m.C())

	return c.UpdateId(m.BID(), bson.M{"$set": values})
}

// Restore using it's ID as search key a model from a persisted MongoDB record.
func RestoreModelByID(m Modeller, id bson.ObjectId) error {
	return RestoreModel(m, bson.M{"_id": id})
}

// Restore a model through any search
func RestoreModel(m Modeller, values bson.M) error {
	c := server.Collection(m.C())

	err := c.Find(values).One(m)
	if err != nil {
		return err
	}

	if empty(m) {
		return errors.New("Could not find that model")
	}

	return nil

}

func setValues(x interface{}, values bson.M) {
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
