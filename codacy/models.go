package codacy

// CodingStandardMeta holds aggregated counts for a coding standard.
type CodingStandardMeta struct {
	EnabledToolsCount       int `json:"enabledToolsCount"`
	EnabledPatternsCount    int `json:"enabledPatternsCount"`
	LinkedRepositoriesCount int `json:"linkedRepositoriesCount"`
}

// CodingStandard represents an organisation coding standard (draft or effective).
type CodingStandard struct {
	ID        int64              `json:"id"`
	Name      string             `json:"name"`
	IsDraft   bool               `json:"isDraft"`
	IsDefault bool               `json:"isDefault"`
	Languages []string           `json:"languages"`
	Meta      CodingStandardMeta `json:"meta"`
}

// CodingStandardResponse wraps a single CodingStandard returned by the API.
type CodingStandardResponse struct {
	Data CodingStandard `json:"data"`
}

// CodingStandardsListResponse wraps a list of CodingStandard values.
type CodingStandardsListResponse struct {
	Data []CodingStandard `json:"data"`
}

// CodingStandardTool represents a tool entry inside a coding standard.
type CodingStandardTool struct {
	CodingStandardID int64  `json:"codingStandardId"`
	UUID             string `json:"uuid"`
	IsEnabled        bool   `json:"isEnabled"`
}

// CodingStandardToolsListResponse wraps a list of CodingStandardTool values.
type CodingStandardToolsListResponse struct {
	Data []CodingStandardTool `json:"data"`
}

// CreateCodingStandardBody is the request body for creating a new coding standard.
type CreateCodingStandardBody struct {
	Name      string   `json:"name"`
	Languages []string `json:"languages"`
}

// UpdatePatternsBody is the request body for the bulk-update patterns endpoint.
type UpdatePatternsBody struct {
	Enabled bool `json:"enabled"`
}

// PromoteResult holds the outcome of promoting a draft coding standard.
type PromoteResult struct {
	Successful []string `json:"successful"`
	Failed     []string `json:"failed"`
}

// PromoteResultResponse wraps PromoteResult returned by the promote endpoint.
type PromoteResultResponse struct {
	Data PromoteResult `json:"data"`
}
