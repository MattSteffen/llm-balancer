package examples

/*
import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	// ...
)

// A struct that will be converted to a Structured Outputs response schema
type HistoricalComputer struct {
	Origin       Origin   `json:"origin" jsonschema_description:"The origin of the computer"`
	Name         string   `json:"full_name" jsonschema_description:"The name of the device model"`
	Legacy       string   `json:"legacy" jsonschema:"enum=positive,enum=neutral,enum=negative" jsonschema_description:"Its influence on the field of computing"`
	NotableFacts []string `json:"notable_facts" jsonschema_description:"A few key facts about the computer"`
}

type Origin struct {
	YearBuilt    int64  `json:"year_of_construction" jsonschema_description:"The year it was made"`
	Organization string `json:"organization" jsonschema_description:"The organization that was in charge of its development"`
}

func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

// Generate the JSON schema at initialization time
var HistoricalComputerResponseSchema = GenerateSchema[HistoricalComputer]()

func main() {

	// ...

	question := "What computer ran the first neural network?"

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "historical_computer",
		Description: openai.String("Notable information about a computer"),
		Schema:      HistoricalComputerResponseSchema,
		Strict:      openai.Bool(true),
	}

	chat, _ := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		// ...
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
		// only certain models can perform structured outputs
		Model: openai.ChatModelGPT4o2024_08_06,
	})

	// extract into a well-typed struct
	var historicalComputer HistoricalComputer
	_ = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &historicalComputer)

	historicalComputer.Name
	historicalComputer.Origin.YearBuilt
	historicalComputer.Origin.Organization
	for i, fact := range historicalComputer.NotableFacts {
		// ...
	}
}
*/
