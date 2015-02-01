package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestSetValues(t *testing.T) {
	type S struct {
		ID       bson.ObjectId `bson:"_id"`
		Username string        `bson:"username"`
	}

	s := S{}
	SetValues(&s, bson.M{"_id": bson.NewObjectId(), "username": "paked"})

	if s.Username != "paked" {
		t.Error("Username should equal `paked`")
	}

	// fmt.Println(s)
}

type Dog struct {
	Model `bson:"model"`

	Name  string `bson:"name"`
	Owner string `bson:"owner"`
	Age   int    `bson:"age"`
}

func NewDog() *Dog {
	dog := &Dog{Model: Model{Collection: "dogs", ID: bson.NewObjectId()}}
	// dog := &Dog{}
	// dog.ID = bson.NewObjectId()
	// dog.Collection = "dogs"

	c := server.Collection(dog.Collection)

	err := c.Insert(dog)
	if err != nil {
		fmt.Println("ERROR INSERTING", err)
		return nil
	}

	return dog
}

func NewDogWithValues(values bson.M) *Dog {
	dog := NewDog()
	err := dog.Update(values)
	if err != nil {
		fmt.Println("ERROR UPDATING MODEL", err)
		// panic(err)
	}
	// fmt.Println(dog, "<--- that is dog equals: ", dog.ID == dog.Model.ID)
	return dog
}

func (d *Dog) GrowOlder() {
	d.Update(bson.M{"age": d.Age + 1})
}

func (d Dog) AgeInDogYears() int {
	return d.Age * 7
}

func TestDog(t *testing.T) {
	server = NewServer()
	d := NewDogWithValues(bson.M{"age": 5, "name": "woof", "owner": "harrison"})
	if d.Age != 5 {
		t.Errorf("Age should be 5 not %v!", d.Age)
	}

	d.GrowOlder()

	if d.Age != 6 {
		t.Errorf("Age should be 6 not %v!", d.Age)
	}

	c := server.Collection(d.Collection)
	e := &Dog{}

	c.Find(bson.M{"_id": d.ID}).One(e)

	if e == (&Dog{}) {
		t.Error("You should have found the dog!")
	}

	// fmt.Println(d, e, User{ID: bson.NewObjectId()})

	d.Delete()
}
