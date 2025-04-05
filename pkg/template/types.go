package template

// metadata from a harness template
type TemplateMetadata struct {
	Name        string                 `json:"name" yaml:"name"`
	Identifier  string                 `json:"identifier" yaml:"identifier"`
	Type        string                 `json:"type" yaml:"type"`
	Description string                 `json:"description" yaml:"description"`
	Author      string                 `json:"author" yaml:"author"`
	Version     string                 `json:"version" yaml:"versionLabel"`
	Tags        []string               `json:"tags" yaml:"tags"`
	Variables   map[string]Variable    `json:"variables" yaml:"variables"`
	Parameters  map[string]Parameter   `json:"parameters" yaml:"parameters"`
	Examples    []string               `json:"examples" yaml:"examples"`
	RawTemplate map[string]interface{} `json:"raw_template"`
}

// template variable
type Variable struct {
	Description string `json:"description" yaml:"description"`
	Type        string `json:"type" yaml:"type"`
	Required    bool   `json:"required" yaml:"required"`
	Scope       string `json:"scope" yaml:"scope"`
}

// template parameter
type Parameter struct {
	Description string      `json:"description" yaml:"description"`
	Type        string      `json:"type" yaml:"type"`
	Required    bool        `json:"required" yaml:"required"`
	Default     interface{} `json:"default" yaml:"default"`
	Scope       string      `json:"scope" yaml:"scope"`
}

// constants for template types
const (
	TemplatePipeline  = "pipeline"
	TemplateStage     = "stage"
	TemplateStepGroup = "stepgroup"
	TemplateStep      = "step"
)

// valid template types
var ValidTemplateTypes = []string{
	TemplatePipeline,
	TemplateStage,
	TemplateStepGroup,
	TemplateStep,
}
