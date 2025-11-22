package handlers

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"

	documentai "cloud.google.com/go/documentai/apiv1"
	"cloud.google.com/go/documentai/apiv1/documentaipb"
	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
	"google.golang.org/api/option"
)

func AdminHandler(c echo.Context) error {
	state := GetStateFromContext(c)

	header := c.Request().Header
	log.Println(header)
	log.Println(header["Hx-Request"] == nil)

	if slices.Contains(services.Admins, state.User.Email) {
		if header["Hx-Request"] == nil {
			return Render(c, http.StatusOK, views.Index(views.Admin(), state))
		} else {
			return Render(c, http.StatusOK, views.Admin())
		}
	} else {
		return c.NoContent(403)
	}
}

func AddIngredient(c echo.Context) error {
	param := c.Param("id")
	id, err := strconv.Atoi(param)
	if err != nil {
		return err
	}

	return Render(c, http.StatusOK, views.Ingredient(id))
}

func AddStep(c echo.Context) error {
	param := c.Param("id")
	id, err := strconv.Atoi(param)
	if err != nil {
		return err
	}

	return Render(c, http.StatusOK, views.Step(id))
}

func DeleteIngredient(c echo.Context) error {
	return c.NoContent(200)
}

func DeleteStep(c echo.Context) error {
	return c.NoContent(200)
}

func ProcessRecipeFile(c echo.Context) error {
	projectID := "r-j-magenta-carrot-42069"
	location := "us"
	// Create a Processor before running sample
	processorID := "a4df3576712f1f6d"
	mimeType := "application/pdf"
	flag.Parse()

	ctx := context.Background()

	endpoint := "us-documentai.googleapis.com:443"
	client, err := documentai.NewDocumentProcessorClient(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		fmt.Println(fmt.Errorf("error creating Document AI client: %w", err))
	}
	defer client.Close()

	// Get file from form input "file"
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to read file")
	}

	req := &documentaipb.ProcessRequest{
		Name: fmt.Sprintf(
			"projects/%s/locations/%s/processors/%s",
			projectID,
			location,
			processorID,
		),
		Source: &documentaipb.ProcessRequest_RawDocument{
			RawDocument: &documentaipb.RawDocument{
				Content:  []byte(data),
				MimeType: mimeType,
			},
		},
	}
	resp, err := client.ProcessDocument(ctx, req)
	if err != nil {
		fmt.Println(fmt.Errorf("processDocument: %w", err))
	}

	// Handle the results.
	document := resp.GetDocument()
	recipe := services.Recipe{}
	for _, e := range document.Entities {
		if e.Type == "Name" {
			recipe.Name = e.MentionText
		}
		if e.Type == "total_time" {
			recipe.Time = e.MentionText
		}
		if e.Type == "Ingredient" {
			recipe.Ingredients = append(recipe.Ingredients, e.MentionText)
		}
		if e.Type == "Step" {
			recipe.Steps = append(recipe.Steps, e.MentionText)
		}
	}

	return Render(c, http.StatusOK, views.AddRecipeForm(recipe))
}
