package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName             = "go-mongo"
	employeeCollection = "employee"
)

type clientWrapper struct {
	*mongo.Client
	employeeCollection *mongo.Collection
}

func newClient(client *mongo.Client, employeeCollection *mongo.Collection) clientWrapper {
	return clientWrapper{
		Client:             client,
		employeeCollection: employeeCollection,
	}
}

type employee struct {
	Name   string `bson:"name"`
	Age    int    `bson:"age"`
	Gender string `bson:"gender"`
}

func connect() (client *mongo.Client, cancel context.CancelFunc, err error) {
	ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
	defer c()
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, nil, err
	}
	return cl, c, nil
}

func (c clientWrapper) insertEmployee(new employee) (insertedId interface{}, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, rErr := c.employeeCollection.InsertOne(ctx, new)
	if rErr != nil {
		return "", rErr
	}
	return r.InsertedID, nil
}

func (c clientWrapper) getEmployees() ([]employee, error) {
	s := make([]employee, 0)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	cursor, err := c.employeeCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var res employee
		err := cursor.Decode(&res)
		if err != nil {
			log.Fatal(err)
		}
		s = append(s, res)
	}
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
	return s, nil
}

func insertPrompt(scanner *bufio.Scanner, client clientWrapper) {
	var name, age, gender string
	fmt.Print("name: ")
	scanner.Scan()
	name = scanner.Text()

	fmt.Print("age: ")
	scanner.Scan()
	age = scanner.Text()
	ageInt, err := strconv.Atoi(age)
	if err != nil {
		log.Println("age must be a value of integer")
		return
	}

	fmt.Print("gender: ")
	scanner.Scan()
	gender = scanner.Text()

	new := employee{
		Name:   name,
		Age:    ageInt,
		Gender: gender,
	}
	id, iErr := client.insertEmployee(new)
	if iErr != nil {
		log.Println("the inputted data is invalid")
		return
	}
	log.Println(id)
}

func insertPromptDebug(
	client clientWrapper,
	name, age, gender string,
) {
	ageInt, err := strconv.Atoi(age)
	if err != nil {
		log.Println("age must be a value of integer")
		return
	}
	new := employee{
		Name:   name,
		Age:    ageInt,
		Gender: gender,
	}
	id, iErr := client.insertEmployee(new)
	if iErr != nil {
		log.Println("the inputted data is invalid")
		return
	}
	log.Println(id)
}

func main() {
	c, cancel, err := connect()
	if err != nil {
		log.Fatal("can't connect to mongo")
	}
	defer cancel()
	client := newClient(c, c.Database(dbName).Collection(employeeCollection))
	defer client.Disconnect(context.Background())
	scanner := bufio.NewScanner(os.Stdin)
	if len(os.Args) > 1 {
		if os.Args[1] == "0" {
			insertPromptDebug(client, os.Args[2], os.Args[3], os.Args[4])
			return
		}
	}
	for {
		scanner.Scan()
		switch scanner.Text() {
		case fmt.Sprint(1):
			s, gErr := client.getEmployees()
			if gErr != nil {
				log.Fatal(gErr)
			}
			if len(s) == 0 {
				log.Println("no employee")
			} else {
				for _, e := range s {
					log.Println(e)
				}
			}
		case fmt.Sprint(2):
			insertPrompt(scanner, client)
		default:
			log.Println("no such option")
		}
	}
}
