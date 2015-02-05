package main

import (
	"gopkg.in/mgo.v2/bson"
	"testing"
)

type Dog struct {
	ID    bson.ObjectId `bson:"_id"`
	Name  string        `bson:"name"`
	Owner string        `bson:"owner"`
	Age   int           `bson:"age"`
}

func (d Dog) BID() bson.ObjectId {
	return d.ID
}

func (d Dog) C() string {
	return "dogs"
}

var d *Dog

func TestModeller(t *testing.T) {
	server = NewServer()
	d = &Dog{ID: bson.NewObjectId(), Name: "Doggy", Owner: "James", Age: 10}

	if err := PersistModel(d); err != nil {
		t.Error("Could not create that model")
		t.FailNow()
	}

	if err := UpdateModel(d, bson.M{"age": 5}); err != nil {
		t.Error("Could not udpate model", err)
		t.FailNow()
	}

	if d.Age != 5 {
		t.Error("Age should be 5, not ", d.Age)
	}
}

func TestPersist(t *testing.T) {
	e := &Dog{}
	err := RestoreModelByID(e, d.BID())

	if err != nil {
		t.Error("Error restoring model:", err)
	}

	if e.BID() != d.BID() {
		t.Error("This is not the same model...")
	}
}

func TestRemove(t *testing.T) {
	if err := RemoveModel(d); err != nil {
		t.Error("Could not remove model:", err)
		t.FailNow()
	}

	e := &Dog{}

	if err := RestoreModelByID(e, d.BID()); err == nil {
		t.Error("Model found.")
	}
}
