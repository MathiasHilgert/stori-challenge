package mailing

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"stori-challenge/internal/summaries"
	"time"

	"github.com/go-gomail/gomail"
)

//go:embed email_template.html
var emailTemplateFS embed.FS

// SMTPConfig holds the configuration for SMTP connection
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPMailer struct {
	config SMTPConfig
}

// NewSMTPMailer creates a new SMTPMailer with the given configuration
func NewSMTPMailer(config SMTPConfig) *SMTPMailer {
	return &SMTPMailer{
		config: config,
	}
}

// Send sends an email with the transaction summary
func (s *SMTPMailer) Send(ctx context.Context, to string, summary summaries.Summary) error {
	// Create a new message
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Resumen de Transacciones - Stori")

	// Generate HTML content
	htmlBody, err := s.generateHTMLBody(summary)
	if err != nil {
		return fmt.Errorf("error generating HTML body: %w", err)
	}

	m.SetBody("text/html", htmlBody)

	// Create SMTP dialer
	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}

// generateHTMLBody generates the HTML body for the email
func (s *SMTPMailer) generateHTMLBody(summary summaries.Summary) (string, error) {
	// Read template from embedded file
	templateContent, err := emailTemplateFS.ReadFile("email_template.html")
	if err != nil {
		return "", fmt.Errorf("error reading email template: %w", err)
	}

	// Create template with custom functions
	t := template.New("email").Funcs(template.FuncMap{
		"monthName": func(month time.Month) string {
			months := map[time.Month]string{
				time.January:   "Enero",
				time.February:  "Febrero",
				time.March:     "Marzo",
				time.April:     "Abril",
				time.May:       "Mayo",
				time.June:      "Junio",
				time.July:      "Julio",
				time.August:    "Agosto",
				time.September: "Septiembre",
				time.October:   "Octubre",
				time.November:  "Noviembre",
				time.December:  "Diciembre",
			}
			return months[month]
		},
		"hasDebit": func(value float64) bool {
			return value != 0.0
		},
		"hasCredit": func(value float64) bool {
			return value != 0.0
		},
		"formatAmount": func(value float64) string {
			return fmt.Sprintf("%.2f", value)
		},
	})

	// Parse template
	t, err = t.Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	// Prepare data for template
	data := struct {
		summaries.Summary
		GeneratedAt string
	}{
		Summary:     summary,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Execute template
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buf.String(), nil
}
