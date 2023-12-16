package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"context"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"google.golang.org/api/option"
)

const firestoreCollection = "characters"

type Character struct {
	ID   int    `json:"id" firestore:"id"`
	Name string `json:"name" firestore:"name"`
}

func newFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	godotenv.Load()
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	projectId := os.Getenv("PROJECT_ID")
	client, err := firestore.NewClient(ctx, projectId, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		log.Fatalf("Erro ao criar cliente Firestore: %v", err)
	}
	return client, err
}

func main() {
	ctx := context.Background()
	client, err := newFirestoreClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{}))
	collection := client.Collection(firestoreCollection)

	e.GET("/characters", func(c echo.Context) error {

		characters, err := collection.Documents(ctx).GetAll()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		var charactersResponse []Character

		for _, doc := range characters {
			var c Character
			if err := doc.DataTo(&c); err != nil {
				log.Printf("Error to convert: %v", err)
				continue
			}
			charactersResponse = append(charactersResponse, c)
		}

		return c.JSON(http.StatusOK, charactersResponse)
	})

	e.GET("/characters/:id", func(c echo.Context) error {
		id := c.Param("id")
		docRef := client.Collection(firestoreCollection).Doc(id)
		snapshot, err := docRef.Get(ctx)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		if !snapshot.Exists() {
			return c.JSON(http.StatusNotFound, nil)
		}

		var character Character
		if err := snapshot.DataTo(&character); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, character)
	})

	e.POST("/characters", func(c echo.Context) error {
		character := new(Character)

		err := c.Bind(character)
		if err != nil {
			c.JSON(http.StatusOK, err)
		}

		docRef := client.Collection(firestoreCollection).Doc(strconv.Itoa(character.ID))
		_, err = docRef.Set(ctx, character)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.NoContent(http.StatusOK)
	})

	e.DELETE("/characters/:id", func(c echo.Context) error {
		id := c.Param("id")
		docRef := client.Collection(firestoreCollection).Doc(id)
		_, err := docRef.Delete(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.NoContent(http.StatusOK)
	})

	e.Start(":8080")
}
