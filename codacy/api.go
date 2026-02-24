package codacy

import (
	"fmt"
	"net/url"
)

// ListCodingStandards returns all coding standards (draft and effective) for an
// organisation.
func (c *Client) ListCodingStandards(provider, orgName string) ([]CodingStandard, error) {
	path := fmt.Sprintf("/organizations/%s/%s/coding-standards", provider, orgName)
	var resp CodingStandardsListResponse
	if err := c.do("GET", path, nil, nil, &resp); err != nil {
		return nil, fmt.Errorf("listCodingStandards: %w", err)
	}
	return resp.Data, nil
}

// GetCodingStandard returns a single coding standard by ID.
func (c *Client) GetCodingStandard(provider, orgName string, id int64) (*CodingStandard, error) {
	path := fmt.Sprintf("/organizations/%s/%s/coding-standards/%d", provider, orgName, id)
	var resp CodingStandardResponse
	if err := c.do("GET", path, nil, nil, &resp); err != nil {
		return nil, fmt.Errorf("getCodingStandard(%d): %w", id, err)
	}
	return &resp.Data, nil
}

// CreateDraftFromStandard creates a new draft coding standard using an existing
// standard as a source (copies its enabled repositories and default status).
// The new draft gets the same name and languages as the source standard.
func (c *Client) CreateDraftFromStandard(provider, orgName string, source CodingStandard) (*CodingStandard, error) {
	path := fmt.Sprintf("/organizations/%s/%s/coding-standards", provider, orgName)

	query := url.Values{}
	query.Set("sourceCodingStandard", fmt.Sprintf("%d", source.ID))

	body := CreateCodingStandardBody{
		Name:      source.Name,
		Languages: source.Languages,
	}

	var resp CodingStandardResponse
	if err := c.do("POST", path, query, body, &resp); err != nil {
		return nil, fmt.Errorf("createDraftFromStandard(%d): %w", source.ID, err)
	}
	return &resp.Data, nil
}

// ListCodingStandardTools returns all tools configured in a coding standard.
func (c *Client) ListCodingStandardTools(provider, orgName string, csID int64) ([]CodingStandardTool, error) {
	path := fmt.Sprintf("/organizations/%s/%s/coding-standards/%d/tools", provider, orgName, csID)
	var resp CodingStandardToolsListResponse
	if err := c.do("GET", path, nil, nil, &resp); err != nil {
		return nil, fmt.Errorf("listCodingStandardTools(%d): %w", csID, err)
	}
	return resp.Data, nil
}

// UpdateSecurityPatterns bulk-enables or bulk-disables all Security-category
// patterns for a specific tool inside a draft coding standard.
func (c *Client) UpdateSecurityPatterns(provider, orgName string, csID int64, toolUUID string, enable bool) error {
	path := fmt.Sprintf(
		"/organizations/%s/%s/coding-standards/%d/tools/%s/patterns/update",
		provider, orgName, csID, toolUUID,
	)
	query := url.Values{}
	query.Set("categories", "Security")

	body := UpdatePatternsBody{Enabled: enable}
	if err := c.do("POST", path, query, body, nil); err != nil {
		return fmt.Errorf("updateSecurityPatterns(cs=%d, tool=%s): %w", csID, toolUUID, err)
	}
	return nil
}

// PromoteDraftCodingStandard promotes a draft coding standard to an effective one.
// The response contains the lists of repositories the standard was successfully (or
// unsuccessfully) applied to.
func (c *Client) PromoteDraftCodingStandard(provider, orgName string, csID int64) (*PromoteResult, error) {
	path := fmt.Sprintf("/organizations/%s/%s/coding-standards/%d/promote", provider, orgName, csID)
	var resp PromoteResultResponse
	if err := c.do("POST", path, nil, nil, &resp); err != nil {
		return nil, fmt.Errorf("promoteDraftCodingStandard(%d): %w", csID, err)
	}
	return &resp.Data, nil
}
