package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ── Core interface ────────────────────────────────────────────────────────────

// Notifier is the primary interface for sending notifications.
type Notifier interface {
	Notify(to, subject, message string) error
}

// Compile-time interface satisfaction checks (nil pointer trick)
var _ Notifier = (*EmailNotifier)(nil)
var _ Notifier = (*SMSNotifier)(nil)
var _ Notifier = (*SlackNotifier)(nil)
var _ Notifier = (*MultiNotifier)(nil)
var _ Notifier = (*RetryNotifier)(nil)
var _ Notifier = (*FilterNotifier)(nil)

// ── Concrete implementations ──────────────────────────────────────────────────

// EmailNotifier sends notifications via (mock) email.
type EmailNotifier struct {
	From string
}

func (e *EmailNotifier) Notify(to, subject, message string) error {
	fmt.Printf("[EMAIL] From:%s To:%s Subject:%q Body:%q\n",
		e.From, to, subject, message)
	return nil
}

// SMSNotifier sends notifications via (mock) SMS.
type SMSNotifier struct {
	Provider string
}

func (s *SMSNotifier) Notify(to, subject, message string) error {
	// SMS ignores subject
	fmt.Printf("[SMS/%s] To:%s Text:%q\n", s.Provider, to, message)
	return nil
}

// SlackNotifier posts notifications to a (mock) Slack channel.
type SlackNotifier struct {
	Webhook string
}

func (sl *SlackNotifier) Notify(to, subject, message string) error {
	fmt.Printf("[SLACK] Channel:%s Subject:%q Message:%q (webhook=%s)\n",
		to, subject, message, sl.Webhook)
	return nil
}

// ── Composed / decorator implementations ─────────────────────────────────────

// MultiNotifier fans out a single Notify call to multiple Notifiers.
// All notifiers are called; any errors are collected and returned together.
type MultiNotifier struct {
	Notifiers []Notifier
}

func (m *MultiNotifier) Notify(to, subject, message string) error {
	var errs []string
	for _, n := range m.Notifiers {
		if err := n.Notify(to, subject, message); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multi-notifier errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// RetryNotifier wraps a Notifier and retries up to MaxRetries times on failure,
// with exponential back-off starting at InitialDelay.
type RetryNotifier struct {
	Inner        Notifier
	MaxRetries   int
	InitialDelay time.Duration
}

func (r *RetryNotifier) Notify(to, subject, message string) error {
	delay := r.InitialDelay
	var lastErr error
	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("[RETRY] attempt %d after %v delay\n", attempt, delay)
			time.Sleep(delay)
			delay *= 2
		}
		if err := r.Inner.Notify(to, subject, message); err != nil {
			lastErr = err
			fmt.Printf("[RETRY] attempt %d failed: %v\n", attempt, err)
			continue
		}
		return nil // success
	}
	return fmt.Errorf("all %d retries exhausted, last error: %w", r.MaxRetries, lastErr)
}

// FilterNotifier wraps a Notifier and applies a predicate before forwarding.
// If the predicate returns false, the message is silently dropped.
type FilterNotifier struct {
	Inner  Notifier
	Filter func(to, subject, message string) bool
}

func (f *FilterNotifier) Notify(to, subject, message string) error {
	if !f.Filter(to, subject, message) {
		fmt.Printf("[FILTER] Dropped message to %s (subject=%q)\n", to, subject)
		return nil
	}
	return f.Inner.Notify(to, subject, message)
}

// ── Nil interface trap demo ───────────────────────────────────────────────────

// failingNotifier is a Notifier whose Notify always returns an error.
// Used to demonstrate the typed nil trap.
type failingNotifier struct{}

func (fn *failingNotifier) Notify(to, subject, message string) error {
	return errors.New("failingNotifier always fails")
}

func demonstrateNilTrap() {
	fmt.Println()
	fmt.Println("=== Nil Interface Trap Demo ===")

	// A typed nil: the interface holds a non-nil type descriptor pointing to a nil value.
	var typed *failingNotifier = nil
	var iface Notifier = typed

	fmt.Printf("typed == nil : %v  (Go compares pointer to nil)\n", typed == nil)
	fmt.Printf("iface == nil : %v  (interface has a non-nil type, so it is NOT nil)\n", iface == nil)

	// Safe check
	if iface != nil {
		fmt.Println("  iface != nil — the interface is non-nil even though the underlying pointer is nil")
		// Calling Notify on a nil *failingNotifier would panic if Notify
		// dereferenced the receiver — guard if needed.
	}

	// Correct way: keep iface as a pure nil interface
	var safeIface Notifier // not assigned — truly nil
	fmt.Printf("safeIface == nil : %v\n", safeIface == nil)
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	fmt.Println("=== Day 31: Go Interfaces Deep Dive ===")
	fmt.Println()

	// Basic notifiers
	email := &EmailNotifier{From: "noreply@example.com"}
	sms := &SMSNotifier{Provider: "Twilio"}
	slack := &SlackNotifier{Webhook: "https://hooks.slack.com/xxx"}

	fmt.Println("--- Single notifiers ---")
	_ = email.Notify("alice@example.com", "Welcome", "Thanks for signing up!")
	_ = sms.Notify("+15555550100", "", "Your OTP is 123456")
	_ = slack.Notify("#alerts", "Deployment", "v1.2.3 deployed to prod")

	// MultiNotifier
	fmt.Println()
	fmt.Println("--- MultiNotifier (fan-out) ---")
	multi := &MultiNotifier{
		Notifiers: []Notifier{email, sms, slack},
	}
	if err := multi.Notify("bob@example.com", "Alert", "Disk usage > 90%"); err != nil {
		fmt.Println("MultiNotifier error:", err)
	}

	// RetryNotifier wrapping a failing notifier
	fmt.Println()
	fmt.Println("--- RetryNotifier (wraps a failing notifier, 3 retries) ---")
	failing := &failingNotifier{}
	retry := &RetryNotifier{
		Inner:        failing,
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
	}
	if err := retry.Notify("carol@example.com", "Test", "Will retry 3 times"); err != nil {
		fmt.Println("RetryNotifier final error:", err)
	}

	// RetryNotifier wrapping a working notifier
	fmt.Println()
	fmt.Println("--- RetryNotifier (wraps a working notifier, succeeds first try) ---")
	retryOK := &RetryNotifier{
		Inner:        email,
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
	}
	_ = retryOK.Notify("dave@example.com", "Good News", "Success!")

	// FilterNotifier
	fmt.Println()
	fmt.Println("--- FilterNotifier (only forward URGENT messages) ---")
	filtered := &FilterNotifier{
		Inner: slack,
		Filter: func(to, subject, message string) bool {
			return strings.Contains(strings.ToUpper(subject), "URGENT")
		},
	}
	_ = filtered.Notify("#oncall", "URGENT: DB down", "Primary DB is unreachable!")
	_ = filtered.Notify("#general", "Weekly report", "Everything is fine.")

	// Nil interface trap
	demonstrateNilTrap()

	fmt.Println()
	fmt.Println("=== Done ===")
}
