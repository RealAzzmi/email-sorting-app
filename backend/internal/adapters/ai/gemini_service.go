package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.Client
}

func NewGeminiService(apiKey string) (*GeminiService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GeminiService{
		client: client,
	}, nil
}

func (g *GeminiService) Close() {
	g.client.Close()
}

func (g *GeminiService) SummarizeEmail(ctx context.Context, email *entities.Email) (string, error) {
	model := g.client.GenerativeModel("gemini-2.0-flash-exp")

	prompt := fmt.Sprintf(`Summarize this email in 1-2 sentences. Be concise and focus on the key action items or main points.

Subject: %s
From: %s
Body: %s

Summary:`, email.Subject, email.Sender, email.Body)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	summary := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return strings.TrimSpace(summary), nil
}

func (g *GeminiService) CategorizeEmail(ctx context.Context, email *entities.Email, categories []entities.Category) ([]int64, error) {
	if len(categories) == 0 {
		return []int64{}, nil
	}

	model := g.client.GenerativeModel("gemini-2.0-flash-exp")

	// Build categories string
	var categoriesStr strings.Builder
	for _, cat := range categories {
		description := ""
		if cat.Description != nil {
			description = *cat.Description
		}
		categoriesStr.WriteString(fmt.Sprintf("- %s: %s\n", cat.Name, description))
	}

	prompt := fmt.Sprintf(`Given the following email and categories, determine which categories this email belongs to. An email can belong to multiple categories or none at all.

Email:
Subject: %s
From: %s
Body: %s

Available Categories:
%s

Instructions:
- Return only the category names that match, separated by commas
- If no categories match, return "NONE"
- Be strict - only categorize if there's a clear match with the category description
- An email can belong to multiple categories

Categories:`, email.Subject, email.Sender, email.Body, categoriesStr.String())

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to categorize email: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return []int64{}, nil
	}

	result := strings.TrimSpace(fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]))

	if result == "NONE" || result == "" {
		return []int64{}, nil
	}

	// Parse the response and match category names to IDs
	categoryNames := strings.Split(result, ",")
	var categoryIDs []int64

	for _, name := range categoryNames {
		name = strings.TrimSpace(name)
		for _, cat := range categories {
			if strings.EqualFold(cat.Name, name) {
				categoryIDs = append(categoryIDs, cat.ID)
				break
			}
		}
	}

	return categoryIDs, nil
}

func (g *GeminiService) AnalyzeUnsubscribePage(ctx context.Context, pageContent, pageURL string) (*repositories.UnsubscribePageAnalysis, error) {
	model := g.client.GenerativeModel("gemini-2.0-flash-exp")

	prompt := fmt.Sprintf(`Analyze this webpage to determine if it's an unsubscribe page and how to interact with it.

URL: %s

Page Content (HTML):
%s

Instructions:
1. Determine if this is an unsubscribe page
2. Identify if authentication/login is required
3. Provide step-by-step actions to unsubscribe (use CSS selectors when possible)
4. Identify success and error indicators
5. Provide reasoning for your analysis

Return your analysis as JSON with this structure:
{
  "is_unsubscribe_page": true/false,
  "requires_auth": true/false,
  "actions": [
    {
      "action": "fill|click|select|wait",
      "selector": "CSS selector or description",
      "value": "text to fill or option to select",
      "reasoning": "why this action is needed"
    }
  ],
  "success_indicators": ["text or selectors that indicate success"],
  "error_indicators": ["text or selectors that indicate errors"],
  "reasoning": "detailed explanation of the page analysis"
}

Actions can be:
- "fill": Fill a text input (requires value)
- "click": Click a button or link
- "select": Select an option from dropdown (requires value)
- "wait": Wait for element to appear

JSON:`, pageURL, pageContent)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to analyze unsubscribe page: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated for page analysis")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Try to parse JSON response
	var analysis repositories.UnsubscribePageAnalysis

	// Clean up the response to extract just the JSON
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return &repositories.UnsubscribePageAnalysis{
			IsUnsubscribePage: false,
			Reasoning:         "Failed to parse AI response as JSON",
		}, nil
	}

	jsonContent := responseText[jsonStart : jsonEnd+1]

	err = json.Unmarshal([]byte(jsonContent), &analysis)
	if err != nil {
		return &repositories.UnsubscribePageAnalysis{
			IsUnsubscribePage: false,
			Reasoning:         fmt.Sprintf("Failed to parse AI response: %v", err),
		}, nil
	}

	return &analysis, nil
}

func (g *GeminiService) ExtractUnsubscribeLink(ctx context.Context, headers, body string) (string, error) {
	model := g.client.GenerativeModel("gemini-2.0-flash-exp")

	prompt := fmt.Sprintf(`Find the unsubscribe link in this email. Return only the URL, nothing else.

Email Headers:
%s

Email Body:
%s

Instructions:
- Look for unsubscribe, opt-out, remove, or email preferences links
- Return the full HTTPS URL only
- If no unsubscribe link found, return "NONE"
- Return only the URL, no explanation

URL:`, headers, body)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to extract unsubscribe link: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	result := strings.TrimSpace(fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]))

	if result == "NONE" || result == "" {
		return "", nil
	}

	// Validate it's a proper HTTPS URL
	if !strings.HasPrefix(result, "https://") {
		return "", nil
	}

	return result, nil
}
