package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go-mongo-replication/db"
	"go-mongo-replication/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

func main() {

	db.Connect()

	app := fiber.New()

	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 30 * time.Second,
	}))

	app.Post("/create-user", createUserHandler)
	app.Get("/get-user/:id", getUserHandler)
	app.Put("/update-user/:id", updateUserHandler)
	app.Delete("/delete-user/:id", deleteUserHandler)

	log.Fatal(app.Listen(":3000"))
}

func createUserHandler(c *fiber.Ctx) error {
	collection := db.Client.Database("test_db").Collection("users")

	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).SendString("Invalid input")
	}

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		fmt.Println(err)
		return c.Status(500).SendString("Failed to insert user")
	}

	return c.JSON(fiber.Map{"id": res.InsertedID})
}

func getUserHandler(c *fiber.Ctx) error {
	collection := db.Client.Database("test_db").Collection("users")

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user model.User
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return c.Status(404).SendString("User not found")
	}

	return c.JSON(user)
}

func updateUserHandler(c *fiber.Ctx) error {
	collection := db.Client.Database("test_db").Collection("users")

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	var updates model.User
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).SendString("Invalid input")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates.UpdatedAt = time.Now()
	res, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updates})
	if err != nil {
		return c.Status(500).SendString("Failed to update user")
	}

	if res.MatchedCount == 0 {
		return c.Status(404).SendString("User not found")
	}

	return c.JSON(fiber.Map{"updated": res.ModifiedCount})
}

func deleteUserHandler(c *fiber.Ctx) error {
	collection := db.Client.Database("test_db").Collection("users")

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.Status(500).SendString("Failed to delete user")
	}

	if res.DeletedCount == 0 {
		return c.Status(404).SendString("User not found")
	}

	return c.JSON(fiber.Map{"deleted": res.DeletedCount})
}
