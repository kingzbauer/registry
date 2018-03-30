package main

import (
	"gopkg.in/mgo.v2/bson"
)

// LicenseEntry for a single entry
type LicenseEntry struct {
	LicenceNumber string        `bson:"licence_number" schema:"licence_number"`
	Year          int           `bson:"year" schema:"year"`
	ApplicantName string        `bson:"applicant_commercial_name" schema:"applicant_name"`
	BusinessID    string        `bson:"business_id" schema:"business_id"`
	PinNo         string        `bson:"pin_no" schema:"pin_no"`
	Location      Location      `bson:"location" schema:"location"`
	Category      Category      `bson:"category" schema:"category"`
	ObjectID      bson.ObjectId `bson:"_id,omitempty" schema:"-"`
}

type Location struct {
	Street      string `bson:"street" schema:"street"`
	Town        string `bson:"town" schema:"town"`
	Box         string `bson:"po_box" schema:"box"`
	Building    string `bson:"building" schema:"building"`
	StallNumber string `bson:"stall_number" schema:"stall_number"`
}

type Category string

const (
	TempAlcoholicDrink       Category = "temporary-alcoholic-drinks"
	BrewersAlcoholic         Category = "brewers-alcoholic"
	TravellersAlcoholicDrink Category = "travellers-alcoholic-drinks"
	TheatreAlcoholicDrink    Category = "theatre-alcoholic-drinks"
	ClubAlcoholicDrink       Category = "club-alcoholic-drinks"
	HotelAlcoholicDrink      Category = "Hotel-alcoholic-drinks"
	RestaurantDrink          Category = "restaurant-drinks"
	WholesaleDistributors    Category = "wholesale-distributors"
	GeneralRetails           Category = "general-retails"
	AlcoholRehabilitation    Category = "alcohol-rehabilitation"
)

var CategoryMap = map[string]Category{
	"Temporary Alcoholic Drinks License":  TempAlcoholicDrink,
	"Brewers Alcoholic License":           BrewersAlcoholic,
	"Travellers Alcoholic Drinks License": TravellersAlcoholicDrink,
	"Theatre Alcoholic Drinks License":    TheatreAlcoholicDrink,
	"Club Alcoholic Drinks License":       ClubAlcoholicDrink,
	"Hotel Alcoholic Drinks License":      HotelAlcoholicDrink,
	"Restaurant Drink License":            RestaurantDrink,
	"Wholesale Distributors":              WholesaleDistributors,
	"General Retails":                     GeneralRetails,
	"Alcoholic Rehabilitation":            AlcoholRehabilitation,
}

// Save to the db
func (e *LicenseEntry) Save() error {
	c := db.C("license")
	e.ObjectID = bson.NewObjectId()
	return c.Insert(e)
}

func LicenseByCategory(category Category) ([]*LicenseEntry, error) {
	c := db.C("license")
	q := c.Find(bson.M{"category": category})
	results := make([]*LicenseEntry, 0)
	if err := q.All(&results); err != nil {
		return nil, err
	}
	return results, nil
}
