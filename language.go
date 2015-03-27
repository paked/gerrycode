package main

import (
	"github.com/paked/models"
	"gopkg.in/mgo.v2/bson"
)

type Language struct {
	ID   bson.ObjectId `bson:"_id" json:"id"`
	Name string        `bson:"name" json:"name"`
}

func (l Language) BID() bson.ObjectId {
	return l.ID
}

func (l Language) C() string {
	return "languages"
}

func GetOrCreateLanguage(name string) (Language, error) {
	l := Language{}

	if err := models.Restore(&l, bson.M{"name": name}); err != nil {
		l = Language{ID: bson.NewObjectId(), Name: name}
		if err := models.Persist(l); err != nil {
			return l, err
		}

		return l, nil
	}

	return l, nil
}
