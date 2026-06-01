package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/labstack/echo/v4"

	"github.com/RoastBeefer00/carrot-firebase-server/services"
	"github.com/RoastBeefer00/carrot-firebase-server/views"
)

func AdminHandler(c echo.Context) error {
	state := GetStateFromContext(c)

	header := c.Request().Header
	log.Println(header)
	log.Println(header["Hx-Request"] == nil)

	if slices.Contains(services.Admins, state.User.Email) {
		if header["Hx-Request"] == nil {
			return Render(c, http.StatusOK, views.Index(views.Admin(), state, "admin"))
		} else {
			return Render(c, http.StatusOK, views.Admin())
		}
	} else {
		return c.NoContent(403)
	}
}

func ProcessRecipeFile(c echo.Context) error {
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
	b64 := base64.StdEncoding.EncodeToString(data)

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	var fileBlock anthropic.ContentBlockParamUnion
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		fileBlock = anthropic.NewImageBlockBase64(mimeType, b64)
	default:
		fileBlock = anthropic.NewDocumentBlock(anthropic.Base64PDFSourceParam{Data: b64})
	}

	client := anthropic.NewClient()
	msg, err := client.Messages.New(c.Request().Context(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 4096,
		Tools: []anthropic.ToolUnionParam{
			{
				OfTool: &anthropic.ToolParam{
					Name:        "submit_recipe",
					Description: anthropic.String("Submit the extracted recipe fields."),
					InputSchema: anthropic.ToolInputSchemaParam{
						Properties: map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "Recipe name",
							},
							"time": map[string]any{
								"type":        "string",
								"description": "Total cooking time, e.g. '30 min'",
							},
							"ingredients": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
							"steps": map[string]any{
								"type":  "array",
								"items": map[string]any{"type": "string"},
							},
						},
						Required: []string{"name", "time", "ingredients", "steps"},
					},
				},
			},
		},
		ToolChoice: anthropic.ToolChoiceParamOfTool("submit_recipe"),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				fileBlock,
				anthropic.NewTextBlock("Extract the recipe from this file. Return each ingredient and each step as a separate array element. Use the submit_recipe tool."),
			),
		},
	})
	if err != nil {
		log.Printf("Anthropic API error: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to extract recipe")
	}

	var recipe services.Recipe
	for _, block := range msg.Content {
		if block.Type == "tool_use" {
			if err := json.Unmarshal(block.Input, &recipe); err != nil {
				log.Printf("Failed to unmarshal recipe: %v", err)
				return c.String(http.StatusInternalServerError, "Failed to parse extracted recipe")
			}
			break
		}
	}

	return Render(c, http.StatusOK, views.AddRecipeForm(recipe))
}
