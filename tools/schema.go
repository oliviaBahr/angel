//go:generate go run ./tools/schema.go
//go:build tools

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"angel/src/core"

	"github.com/invopop/jsonschema"
)

func main() {
	// Create a new schema reflector
	r := jsonschema.Reflector{}

	// Generate schema for Config struct
	schema := r.Reflect(&core.Config{})

	// Add custom schema properties
	schema.ID = "https://angel.dev/schema/config.json"
	schema.Title = "Angel Configuration Schema"
	schema.Description = "Configuration schema for the Angel file organizer"

	// Convert to JSON with pretty printing
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	outputFile := "build/angel-config-schema.json"
	err = os.WriteFile(outputFile, schemaJSON, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated JSON schema: %s\n", outputFile)
}
