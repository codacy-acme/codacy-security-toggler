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

// ListRepositoriesWithAnalysis returns all repositories for an organisation,
// following cursor-based pagination automatically.
func (c *Client) ListRepositoriesWithAnalysis(provider, orgName string) ([]RepositoryWithAnalysis, error) {
	path := fmt.Sprintf("/analysis/organizations/%s/%s/repositories", provider, orgName)
	var all []RepositoryWithAnalysis
	cursor := ""
	for {
		query := url.Values{}
		query.Set("limit", "100")
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		var resp RepositoryWithAnalysisListResponse
		if err := c.do("GET", path, query, nil, &resp); err != nil {
			return nil, fmt.Errorf("listRepositoriesWithAnalysis: %w", err)
		}
		all = append(all, resp.Data...)
		if resp.Pagination == nil || resp.Pagination.Cursor == "" {
			break
		}
		cursor = resp.Pagination.Cursor
	}
	return all, nil
}

// ListRepositoryTools returns the analysis tools configured for a repository.
func (c *Client) ListRepositoryTools(provider, orgName, repoName string) ([]AnalysisTool, error) {
	path := fmt.Sprintf("/analysis/organizations/%s/%s/repositories/%s/tools",
		provider, orgName, repoName)
	var resp AnalysisToolsListResponse
	if err := c.do("GET", path, nil, nil, &resp); err != nil {
		return nil, fmt.Errorf("listRepositoryTools(%s): %w", repoName, err)
	}
	return resp.Data, nil
}

// UpdateRepositorySecurityPatterns bulk-enables or bulk-disables all
// Security-category patterns for a specific tool in a repository.
// Uses PATCH /analysis/.../tools/{toolUuid}/patterns?categories=Security.
func (c *Client) UpdateRepositorySecurityPatterns(provider, orgName, repoName, toolUUID string, enable bool) error {
	path := fmt.Sprintf("/analysis/organizations/%s/%s/repositories/%s/tools/%s/patterns",
		provider, orgName, repoName, toolUUID)
	query := url.Values{}
	query.Set("categories", "Security")
	body := UpdatePatternsBody{Enabled: enable}
	if err := c.do("PATCH", path, query, body, nil); err != nil {
		return fmt.Errorf("updateRepositorySecurityPatterns(repo=%s, tool=%s): %w", repoName, toolUUID, err)
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
