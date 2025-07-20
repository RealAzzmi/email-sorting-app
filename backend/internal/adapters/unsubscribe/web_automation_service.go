package unsubscribe

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/email-sorting-app/internal/domain/entities"
	"github.com/email-sorting-app/internal/domain/repositories"
	"github.com/playwright-community/playwright-go"
)

type WebAutomationService struct {
	aiService  repositories.AIService
	timeout    time.Duration
	playwright *playwright.Playwright
	browser    playwright.Browser
}

func NewWebAutomationService(aiService repositories.AIService) *WebAutomationService {
	return &WebAutomationService{
		aiService: aiService,
		timeout:   30 * time.Second,
	}
}

func (w *WebAutomationService) initBrowser(ctx context.Context) error {
	if w.playwright == nil {
		pw, err := playwright.Run()
		if err != nil {
			return fmt.Errorf("failed to run playwright: %w", err)
		}
		w.playwright = pw
	}

	if w.browser == nil {
		browser, err := w.playwright.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
				"--disable-accelerated-2d-canvas",
				"--no-first-run",
				"--no-zygote",
				"--disable-gpu",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to launch browser: %w", err)
		}
		w.browser = browser
	}

	return nil
}

func (w *WebAutomationService) Close() error {
	if w.browser != nil {
		if err := w.browser.Close(); err != nil {
			return fmt.Errorf("failed to close browser: %w", err)
		}
	}
	if w.playwright != nil {
		if err := w.playwright.Stop(); err != nil {
			return fmt.Errorf("failed to stop playwright: %w", err)
		}
	}
	return nil
}

func (w *WebAutomationService) UnsubscribeFromEmail(ctx context.Context, email *entities.Email) (*repositories.UnsubscribeResult, error) {	
	if email.UnsubscribeLink == nil || *email.UnsubscribeLink == "" {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   "No unsubscribe link found in email",
			ErrorType: "no_link",
		}, nil
	}

	link := *email.UnsubscribeLink
	
	// Validate the link before processing
	if !w.isValidUnsubscribeLink(link) {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   "Invalid or suspicious unsubscribe link",
			ErrorType: "invalid_link",
		}, nil
	}

	// For now, implement a simplified approach that handles common unsubscribe patterns
	// This can be extended with actual browser automation using Playwright
	result, err := w.processUnsubscribeLink(ctx, link, email)
	if err != nil {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to process unsubscribe: %v", err),
			ErrorType: "processing_error",
		}, nil
	}
	return result, nil
}

func (w *WebAutomationService) BulkUnsubscribe(ctx context.Context, emails []*entities.Email) (map[int64]*repositories.UnsubscribeResult, error) {
	results := make(map[int64]*repositories.UnsubscribeResult)
	
	for _, email := range emails {
		// Add delay between requests to be respectful
		time.Sleep(2 * time.Second)
		
		result, err := w.UnsubscribeFromEmail(ctx, email)
		if err != nil {
			results[email.ID] = &repositories.UnsubscribeResult{
				Success:   false,
				Message:   fmt.Sprintf("Error: %v", err),
				ErrorType: "processing_error",
			}
		} else {
			results[email.ID] = result
		}
	}
	
	return results, nil
}

func (w *WebAutomationService) ValidateUnsubscribeLink(ctx context.Context, link string) (bool, error) {
	return w.isValidUnsubscribeLink(link), nil
}

func (w *WebAutomationService) isValidUnsubscribeLink(link string) bool {
	// Basic validation for unsubscribe links
	if link == "" {
		return false
	}
	
	// Must be HTTPS for security
	if !strings.HasPrefix(link, "https://") {
		return false
	}
	
	// Common unsubscribe patterns
	unsubscribePatterns := []string{
		"unsubscribe",
		"opt-out",
		"remove",
		"list-unsubscribe",
		"email-preferences",
		"subscription",
	}
	
	linkLower := strings.ToLower(link)
	for _, pattern := range unsubscribePatterns {
		if strings.Contains(linkLower, pattern) {
			return true
		}
	}
	
	return false
}

func (w *WebAutomationService) processUnsubscribeLink(ctx context.Context, link string, email *entities.Email) (*repositories.UnsubscribeResult, error) {
	// Initialize browser if not already done
	if err := w.initBrowser(ctx); err != nil {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to initialize browser: %v", err),
			ErrorType: "browser_init_error",
		}, nil
	}

	// Create a new page
	page, err := w.browser.NewPage()
	if err != nil {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to create new page: %v", err),
			ErrorType: "page_creation_error",
		}, nil
	}
	defer page.Close()

	// Set page timeout
	page.SetDefaultTimeout(float64(w.timeout.Milliseconds()))

	// Navigate to the unsubscribe page
	_, err = page.Goto(link, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to navigate to page: %v", err),
			ErrorType: "navigation_error",
		}, nil
	}

	// Wait for page to fully load
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		// Don't fail immediately, continue with partial load
		fmt.Printf("Warning: page didn't reach networkidle state: %v\n", err)
	}

	// Get page content for AI analysis
	content, err := page.Content()
	if err != nil {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to get page content: %v", err),
			ErrorType: "content_extraction_error",
		}, nil
	}

	// Use AI to analyze the page and determine actions
	analysis, err := w.aiService.AnalyzeUnsubscribePage(ctx, content, link)
	if err != nil {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   fmt.Sprintf("Failed to analyze page with AI: %v", err),
			ErrorType: "ai_analysis_error",
		}, nil
	}

	// Handle authentication requirement
	if analysis.RequiresAuth {
		return &repositories.UnsubscribeResult{
			Success:      false,
			Message:      "Page requires authentication to unsubscribe",
			ErrorType:    "auth_required",
			RequiresAuth: true,
		}, nil
	}

	// If AI indicates it's not an unsubscribe page
	if !analysis.IsUnsubscribePage {
		return &repositories.UnsubscribeResult{
			Success:   false,
			Message:   "Page does not appear to be a valid unsubscribe page",
			ErrorType: "invalid_page",
		}, nil
	}

	// Execute the actions determined by AI
	for i, action := range analysis.Actions {
		if err := w.executeAction(page, action); err != nil {
			return &repositories.UnsubscribeResult{
				Success:   false,
				Message:   fmt.Sprintf("Failed to execute action %d (%s): %v", i+1, action.Action, err),
				ErrorType: "action_execution_error",
			}, nil
		}

		// Small delay between actions
		time.Sleep(500 * time.Millisecond)
	}

	// Verify success if possible
	successVerified := w.verifyUnsubscribeSuccess(page)
	
	if successVerified {
		return &repositories.UnsubscribeResult{
			Success: true,
			Message: fmt.Sprintf("Successfully unsubscribed from %s", email.Sender),
		}, nil
	} else {
		return &repositories.UnsubscribeResult{
			Success: true, // Still consider it successful if actions completed
			Message: fmt.Sprintf("Unsubscribe actions completed for %s (verification unclear)", email.Sender),
		}, nil
	}
}

func (w *WebAutomationService) executeAction(page playwright.Page, action repositories.UnsubscribeAction) error {
	switch action.Action {
	case "click":
		element, err := page.QuerySelector(action.Selector)
		if err != nil || element == nil {
			// Try alternative selectors or text-based selection
			if strings.Contains(action.Selector, "button") || strings.Contains(action.Value, "unsubscribe") {
				// Try finding button by text
				elements, _ := page.QuerySelectorAll("button, input[type='button'], input[type='submit'], a")
				for _, el := range elements {
					text, _ := el.TextContent()
					if strings.Contains(strings.ToLower(text), "unsubscribe") || 
					   strings.Contains(strings.ToLower(text), "remove") ||
					   strings.Contains(strings.ToLower(text), "opt") {
						return el.Click()
					}
				}
			}
			return fmt.Errorf("element not found: %s", action.Selector)
		}
		return element.Click()

	case "fill":
		element, err := page.QuerySelector(action.Selector)
		if err != nil || element == nil {
			return fmt.Errorf("input element not found: %s", action.Selector)
		}
		return element.Fill(action.Value)

	case "select":
		_, err := page.SelectOption(action.Selector, playwright.SelectOptionValues{
			Values: &[]string{action.Value},
		})
		return err

	case "wait":
		duration, err := time.ParseDuration(action.Value)
		if err != nil {
			duration = 2 * time.Second // Default wait
		}
		time.Sleep(duration)
		return nil

	case "submit":
		// Find and submit the form
		if action.Selector != "" {
			element, err := page.QuerySelector(action.Selector)
			if err != nil || element == nil {
				return fmt.Errorf("form not found: %s", action.Selector)
			}
			return element.Click() // Usually submit button
		} else {
			// Find any submit button
			submitBtn, err := page.QuerySelector("input[type='submit'], button[type='submit']")
			if err == nil && submitBtn != nil {
				return submitBtn.Click()
			}
			return fmt.Errorf("no submit button found")
		}

	default:
		return fmt.Errorf("unknown action: %s", action.Action)
	}
}

func (w *WebAutomationService) verifyUnsubscribeSuccess(page playwright.Page) bool {
	// Look for common success indicators
	successIndicators := []string{
		"successfully unsubscribed",
		"you have been unsubscribed",
		"removed from our mailing list",
		"unsubscribe successful",
		"email preferences updated",
		"subscription cancelled",
	}

	content, err := page.TextContent("body")
	if err != nil {
		return false
	}

	contentLower := strings.ToLower(content)
	for _, indicator := range successIndicators {
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}

	// Check for success-related CSS classes or IDs
	successElements, _ := page.QuerySelectorAll(".success, .confirmation, #success, #confirmation, .alert-success")
	return len(successElements) > 0
}

// ExtractUnsubscribeLink extracts unsubscribe links from email content
func ExtractUnsubscribeLink(headers, body string) *string {
	// Check List-Unsubscribe header first (RFC 2369)
	listUnsubscribeRegex := regexp.MustCompile(`List-Unsubscribe:\s*<([^>]+)>`)
	if matches := listUnsubscribeRegex.FindStringSubmatch(headers); len(matches) > 1 {
		link := strings.TrimSpace(matches[1])
		if strings.HasPrefix(link, "https://") {
			return &link
		}
	}
	
	// Look for unsubscribe links in email body
	unsubscribeRegexes := []*regexp.Regexp{
		regexp.MustCompile(`href=["']([^"']*unsubscribe[^"']*)["']`),
		regexp.MustCompile(`href=["']([^"']*opt-out[^"']*)["']`),
		regexp.MustCompile(`href=["']([^"']*remove[^"']*)["']`),
		regexp.MustCompile(`href=["']([^"']*email-preferences[^"']*)["']`),
		regexp.MustCompile(`href=["']([^"']*subscription[^"']*)["']`),
	}
	
	for _, regex := range unsubscribeRegexes {
		if matches := regex.FindStringSubmatch(body); len(matches) > 1 {
			link := strings.TrimSpace(matches[1])
			// Ensure it's a valid HTTPS URL
			if strings.HasPrefix(link, "https://") {
				return &link
			}
			// Handle relative URLs or protocol-relative URLs
			if strings.HasPrefix(link, "//") {
				httpsLink := "https:" + link
				return &httpsLink
			}
		}
	}
	
	return nil
}