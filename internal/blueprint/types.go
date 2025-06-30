package blueprint

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/jsonschema"
)

// Token represents a part of a shell word after parsing
type Token interface {
	String() string
}

// TextToken represents literal text in a shell word
type TextToken struct {
	Value string
}

func (t TextToken) String() string {
	return t.Value
}

// FieldToken represents a template field in a shell word
type FieldToken struct {
	Name         string
	Description  string
	Required     bool
	OriginalFlag string // For boolean flags, stores the original flag format (e.g., "-f", "--verbose")
	OriginalName string // For templates, stores the original name with spacing (e.g., "page " from "{{page # desc}}")
}

func (t FieldToken) String() string {
	if t.Required {
		return "{{" + t.Name + "}}"
	}
	return "[" + t.Name + "]"
}

// Blueprint represents a parsed command template
type Blueprint struct {
	BaseCommand string
	InputSchema *jsonschema.Schema
	ShellWords  [][]Token // Tokenized shell words
}

// GetBaseCommand returns the base command
func (bp *Blueprint) GetBaseCommand() string {
	return bp.BaseCommand
}

// GetCommandFormat returns the command format without the "Run the shell command" prefix
func (bp *Blueprint) GetCommandFormat() string {
	parts := make([]string, len(bp.ShellWords))
	for i, tokens := range bp.ShellWords {
		parts[i] = bp.renderTokensForDisplay(tokens)
	}
	return strings.Join(parts, " ")
}

// renderTokensForDisplay renders tokens for display purposes (used in command format)
func (bp *Blueprint) renderTokensForDisplay(tokens []Token) string {
	if len(tokens) == 1 {
		if fieldToken, ok := tokens[0].(FieldToken); ok {
			return bp.renderFieldTokenForDisplay(fieldToken)
		}
		return tokens[0].String()
	}

	var result strings.Builder
	for _, token := range tokens {
		switch t := token.(type) {
		case TextToken:
			result.WriteString(t.Value)
		case FieldToken:
			result.WriteString(bp.renderFieldTokenForDisplay(t))
		}
	}
	return result.String()
}

// renderFieldTokenForDisplay renders a single field token for display
func (bp *Blueprint) renderFieldTokenForDisplay(token FieldToken) string {
	// For boolean flags, use the original flag format
	if token.OriginalFlag != "" {
		return "[" + token.OriginalFlag + "]"
	}

	// For required fields, use template format
	if token.Required {
		// Use OriginalName if available (preserves spacing), otherwise use Name
		name := token.Name
		if token.OriginalName != "" {
			name = token.OriginalName
		}
		return "{{" + name + "}}"
	}

	// For optional fields, check if it's an array by looking at the schema
	normalizedName := strings.ReplaceAll(token.Name, "-", "_")
	if schema, exists := bp.InputSchema.Properties[normalizedName]; exists && schema.Type == "array" {
		return "[" + normalizedName + "...]"
	}

	// Regular optional field
	return "[" + normalizedName + "]"
}

// GetInputSchema returns the input schema
func (bp *Blueprint) GetInputSchema() interface{} {
	return bp.InputSchema
}
