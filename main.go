package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codacy/codacy-security-toggler/codacy"
)

func main() {
	var (
		apiToken = flag.String("api-token", "", "Codacy API token (or set CODACY_API_TOKEN)")
		provider = flag.String("provider", "gh", "Git provider: gh (GitHub), gl (GitLab), bb (Bitbucket)")
		orgName  = flag.String("organization", "", "Organisation name on the Git provider (required)")
		csID     = flag.Int64("coding-standard-id", 0, "ID of the coding standard to process (0 = all standards)")
		enable   = flag.Bool("enable", true, "true = enable security patterns, false = disable them")
		promote  = flag.Bool("promote", true, "Promote the draft after updating patterns")
		skipLive = flag.Bool("skip-live", false, "Skip coding standards that are not drafts (instead of duplicating them)")
		dryRun   = flag.Bool("dry-run", false, "Print what would happen without making any changes")
		verbose  = flag.Bool("verbose", false, "Print additional detail (tool UUIDs, etc.)")
	)
	flag.Usage = usage
	flag.Parse()

	token := *apiToken
	if token == "" {
		token = os.Getenv("CODACY_API_TOKEN")
	}
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: API token is required — use --api-token or set CODACY_API_TOKEN")
		flag.Usage()
		os.Exit(1)
	}
	if *orgName == "" {
		fmt.Fprintln(os.Stderr, "error: --organization is required")
		flag.Usage()
		os.Exit(1)
	}

	action := "enable"
	if !*enable {
		action = "disable"
	}

	fmt.Println("Codacy Security Pattern Toggler")
	fmt.Printf("  Provider:     %s\n", *provider)
	fmt.Printf("  Organisation: %s\n", *orgName)
	fmt.Printf("  Action:       %s security patterns\n", action)
	fmt.Printf("  Promote:      %v\n", *promote)
	if *dryRun {
		fmt.Println("  Mode:         DRY RUN (no changes will be made)")
	}
	fmt.Println()

	client := codacy.NewClient(token)

	standards, err := resolveStandards(client, *provider, *orgName, *csID)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if len(standards) == 0 {
		fmt.Println("No coding standards found.")
		return
	}

	fmt.Printf("Found %d coding standard(s) to process:\n", len(standards))
	for _, cs := range standards {
		fmt.Printf("  [%d] %s  (draft=%v  default=%v  tools=%d  patterns=%d)\n",
			cs.ID, cs.Name, cs.IsDraft, cs.IsDefault,
			cs.Meta.EnabledToolsCount, cs.Meta.EnabledPatternsCount)
	}
	fmt.Println()

	var hadError bool
	for _, cs := range standards {
		if err := processStandard(client, *provider, *orgName, cs, *enable, *promote, *skipLive, *dryRun, *verbose); err != nil {
			log.Printf("error processing %q (ID %d): %v", cs.Name, cs.ID, err)
			hadError = true
		}
	}

	// Phase 2: repositories not covered by any coding standard.
	fmt.Println("--- Detached repositories (not following any coding standard) ---")
	fmt.Println()
	if err := processDetachedRepositories(client, *provider, *orgName, *enable, *dryRun, *verbose); err != nil {
		log.Printf("error processing detached repositories: %v", err)
		hadError = true
	}

	if hadError {
		os.Exit(1)
	}
}

// resolveStandards returns the list of coding standards to operate on.
// When id > 0 it fetches that single standard; otherwise it lists all standards.
func resolveStandards(client *codacy.Client, provider, orgName string, id int64) ([]codacy.CodingStandard, error) {
	if id != 0 {
		cs, err := client.GetCodingStandard(provider, orgName, id)
		if err != nil {
			return nil, err
		}
		return []codacy.CodingStandard{*cs}, nil
	}
	return client.ListCodingStandards(provider, orgName)
}

// processStandard runs the full toggle-and-promote workflow for one coding standard.
func processStandard(
	client *codacy.Client,
	provider, orgName string,
	cs codacy.CodingStandard,
	enable, promote, skipLive, dryRun, verbose bool,
) error {
	fmt.Printf("==> %q (ID %d)\n", cs.Name, cs.ID)

	target := cs

	// Non-draft standards require a new draft to be created before they can be edited.
	if !cs.IsDraft {
		if skipLive {
			fmt.Println("    Skipping — standard is not a draft and --skip-live is set")
			fmt.Println()
			return nil
		}
		fmt.Println("    Standard is not a draft — creating a draft from it…")
		if !dryRun {
			dup, err := client.CreateDraftFromStandard(provider, orgName, cs)
			if err != nil {
				return fmt.Errorf("creating draft from standard: %w", err)
			}
			target = *dup
			fmt.Printf("    Draft created: %q (ID %d)\n", target.Name, target.ID)
		} else {
			fmt.Printf("    [dry-run] would create a draft from standard %d\n", cs.ID)
		}
	}

	// List all tools in the (draft) coding standard.
	tools, err := client.ListCodingStandardTools(provider, orgName, target.ID)
	if err != nil {
		return fmt.Errorf("listing tools: %w", err)
	}
	fmt.Printf("    Tools found: %d\n", len(tools))

	action := "Enabling"
	if !enable {
		action = "Disabling"
	}

	// Bulk-update security patterns per tool.
	var failedTools []string
	for _, tool := range tools {
		if verbose {
			fmt.Printf("    %s security patterns for tool %s\n", action, tool.UUID)
		}
		if !dryRun {
			if err := client.UpdateSecurityPatterns(provider, orgName, target.ID, tool.UUID, enable); err != nil {
				log.Printf("    warning: could not update tool %s: %v", tool.UUID, err)
				failedTools = append(failedTools, tool.UUID)
				continue
			}
		} else {
			fmt.Printf("    [dry-run] would %s security patterns for tool %s\n",
				strings.ToLower(action), tool.UUID)
		}
	}

	updated := len(tools) - len(failedTools)
	fmt.Printf("    %s security patterns: %d/%d tool(s) updated\n", action, updated, len(tools))
	if len(failedTools) > 0 {
		fmt.Printf("    Failed tools: %s\n", strings.Join(failedTools, ", "))
	}

	// Promote the draft to an effective coding standard.
	if promote {
		fmt.Println("    Promoting draft…")
		if !dryRun {
			result, err := client.PromoteDraftCodingStandard(provider, orgName, target.ID)
			if err != nil {
				return fmt.Errorf("promoting standard: %w", err)
			}
			fmt.Println("    Promoted successfully!")
			if len(result.Successful) > 0 {
				fmt.Printf("    Applied to %d repo(s): %s\n",
					len(result.Successful), strings.Join(result.Successful, ", "))
			}
			if len(result.Failed) > 0 {
				fmt.Printf("    Failed for %d repo(s): %s\n",
					len(result.Failed), strings.Join(result.Failed, ", "))
			}
		} else {
			fmt.Printf("    [dry-run] would promote draft standard %d\n", target.ID)
		}
	}

	fmt.Println()
	return nil
}

// processDetachedRepositories handles repositories that are not covered by any
// coding standard by toggling their Security-category patterns directly.
func processDetachedRepositories(client *codacy.Client, provider, orgName string, enable, dryRun, verbose bool) error {
	repos, err := client.ListRepositoriesWithAnalysis(provider, orgName)
	if err != nil {
		return fmt.Errorf("listing repositories: %w", err)
	}

	var detached []codacy.RepositoryWithAnalysis
	for _, r := range repos {
		if len(r.Repository.Standards) == 0 {
			detached = append(detached, r)
		}
	}

	if len(detached) == 0 {
		fmt.Println("No detached repositories found.")
		fmt.Println()
		return nil
	}

	fmt.Printf("Found %d detached repository(ies):\n", len(detached))
	for _, r := range detached {
		fmt.Printf("  - %s\n", r.Repository.Name)
	}
	fmt.Println()

	action := "Enabling"
	if !enable {
		action = "Disabling"
	}

	var hadError bool
	for _, r := range detached {
		repoName := r.Repository.Name
		fmt.Printf("==> %s\n", repoName)

		tools, err := client.ListRepositoryTools(provider, orgName, repoName)
		if err != nil {
			log.Printf("    error listing tools for %s: %v", repoName, err)
			hadError = true
			continue
		}
		fmt.Printf("    Tools found: %d\n", len(tools))

		var failedTools []string
		for _, tool := range tools {
			if verbose {
				fmt.Printf("    %s security patterns for tool %s (%s)\n", action, tool.Name, tool.UUID)
			}
			if !dryRun {
				if err := client.UpdateRepositorySecurityPatterns(provider, orgName, repoName, tool.UUID, enable); err != nil {
					log.Printf("    warning: could not update tool %s: %v", tool.UUID, err)
					failedTools = append(failedTools, tool.UUID)
					continue
				}
			} else {
				fmt.Printf("    [dry-run] would %s security patterns for tool %s (%s)\n",
					strings.ToLower(action), tool.Name, tool.UUID)
			}
		}

		updated := len(tools) - len(failedTools)
		fmt.Printf("    %s security patterns: %d/%d tool(s) updated\n", action, updated, len(tools))
		if len(failedTools) > 0 {
			fmt.Printf("    Failed tools: %s\n", strings.Join(failedTools, ", "))
		}
		fmt.Println()
	}

	if hadError {
		return fmt.Errorf("one or more detached repositories could not be fully updated")
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: codacy-security-toggler [flags]

Toggles Security-category code patterns across all tools of one or more
coding standards in a Codacy organisation, then optionally promotes the
updated draft to an effective coding standard.

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:

  # Enable security patterns on all coding standards and promote each draft
  codacy-security-toggler \
    --api-token=$CODACY_API_TOKEN \
    --provider=gh \
    --organization=my-org \
    --enable=true

  # Disable security patterns on a specific coding standard (dry run first)
  codacy-security-toggler \
    --api-token=$CODACY_API_TOKEN \
    --organization=my-org \
    --coding-standard-id=42 \
    --enable=false \
    --dry-run

  # Enable without promoting (leave as draft for review)
  codacy-security-toggler \
    --api-token=$CODACY_API_TOKEN \
    --organization=my-org \
    --enable=true \
    --promote=false
`)
}
