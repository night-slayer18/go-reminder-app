package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			fmt.Println("Error loading .env file")
			log.Fatal("Error loading .env file", err)
		}
	}
	MONGODB_URL := os.Getenv("MONGO_URL")
	clientOptions := options.Client().ApplyURI(MONGODB_URL)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection = client.Database("reminder").Collection("todos")

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Delete("/api/todos/:id", deleteTodo)
	app.Patch("/api/todos/:id", updateTodo)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
		app.Get("*", func(c *fiber.Ctx) error {
			return c.SendFile("./client/build/index.html")
		})
	}

	fmt.Println("Server is running on port", PORT)
	log.Fatal(app.Listen(":" + PORT))
}

func getTodos(c *fiber.Ctx) error {
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
		return err
	}
	var todos []Todo
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			log.Fatal(err)
			return err
		}
		todos = append(todos, todo)
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	if err := c.BodyParser(todo); err != nil {
		log.Fatal(err)
		return c.Status(400).JSON(fiber.Map{"msg": "Invalid request"})
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "Body is required"})
	}

	result, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		log.Fatal(err)
		return err
	}

	todo.ID = result.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
		return c.Status(400).JSON(fiber.Map{"msg": "Invalid ID"})
	}

	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		log.Fatal(err)
		return err
	}
	return c.Status(200).JSON(fiber.Map{"success": true, "msg": "Todo deleted"})
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
		return c.Status(400).JSON(fiber.Map{"msg": "Invalid ID"})
	}

	result := collection.FindOne(context.Background(), bson.M{"_id": objectID})

	var todo Todo
	err = result.Decode(&todo)
	if err != nil {
		log.Fatal(err)
		return err
	}

	todo.Completed = true

	_, err = collection.ReplaceOne(context.Background(), bson.M{"_id": objectID}, todo)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true, "msg": "Todo updated"})
}
