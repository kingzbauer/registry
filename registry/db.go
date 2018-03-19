package main

import (
	"io/ioutil"
	"mime/multipart"

	"fmt"
	riak "github.com/basho/riak-go-client"
	util "github.com/basho/taste-of-riak/go/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

var (
	session *mgo.Session
	db      *mgo.Database
	riakC   *riak.Client
)

type Entry struct {
	Name string        `bson:"name"`
	Age  string        `bson:"age"`
	ID   bson.ObjectId `bson:"_id"`
}

type File struct {
	Name        string
	Content     []byte
	ContentType string
}

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

func closeDB() {
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
		ContentType: fileHeader.Header.Get("Content-Type"),
		Value:       data,
		UserMeta: []*riak.Pair{
			{Key: "filename", Value: fileHeader.Header.Get("name")},
		},
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

func list() ([]*Entry, error) {
	c := db.C("entry")
	q := c.Find(nil)
	entries := make([]*Entry, 0)
	if err := q.All(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func get(id string) (*Entry, error) {
	c := db.C("entry")
	q := c.FindId(bson.ObjectIdHex(id))
	e := new(Entry)
	if err := q.One(e); err != nil {
		return nil, err
	}
	return e, nil
}

func getFile(id string) (*File, error) {
	var (
		cmd riak.Command
		err error
	)

	if cmd, err = riak.NewFetchValueCommandBuilder().
		WithBucket("registry").
		WithKey(id).
		Build(); err != nil {
		return nil, err
	}
	done := make(chan riak.Command, 1)
	wg := &sync.WaitGroup{}
	a := &riak.Async{
		Command: cmd,
		Wait:    wg,
		Done:    done,
	}
	if err = riakC.ExecuteAsync(a); err != nil {
		return nil, err
	}
	wg.Wait()
	close(done)

	for d := range done {
		f := d.(*riak.FetchValueCommand)
		if f.Response.IsNotFound {
			return nil, nil
		}

		obj := f.Response.Values[0]
		file := &File{}
		if !obj.HasUserMeta() {
			file.Name = id
		} else {
			file.Name = obj.UserMeta[0].Value
		}
		file.Content = obj.Value
		file.ContentType = obj.ContentType
		return file, nil
	}
	return nil, nil
}
