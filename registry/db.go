package main

import (
	"mime/multipart"

	riak "github.com/basho/riak-go-client"
	util "github.com/basho/taste-of-riak/go/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
)

var (
	session *mgo.Session
	db      *mgo.Database
	riakC   *riak.Client
)

func init() {
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	db = session.DB("registry")

	// initial riak
	o := &riak.NewClientOptions{
		RemoteAddresses: []string{"localhost:8087"},
	}
	riakC, err = riak.NewClient(o)
	if err != nil {
		util.ErrExit(err)
	}
}

func close() {
	if session != nil {
		session.Close()
	}

	if riakC != nil {
		if err := riakC.Stop(); err != nil {
			util.ErrExit(err)
		}
	}
}

func newEntry(name, age string, fileHeader *multipart.FileHeader) (bson.ObjectId, error) {
	c := db.C("entry")
	id := bson.NewObjectId()
	doc := bson.M{
		"name": name,
		"age":  age,
		"_id":  id,
	}
	if err := c.Insert(doc); err != nil {
		return "", err
	}

	// if a file is present, retrieve it
	if fileHeader == nil {
		return id, nil
	}

	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	// insert the file into riak
	obj := &riak.Object{
		Bucket:      "registry",
		Key:         id.String(),
		ContentType: "application/octet-stream",
		Value:       data,
	}

	cmd, err := riak.NewStoreValueCommandBuilder().
		WithContent(obj).
		Build()
	err = riakC.Execute(cmd)
	if err != nil {
		return "", err
	}
	return id, nil
}
