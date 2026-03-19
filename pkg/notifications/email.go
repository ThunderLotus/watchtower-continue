package notifications

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	t "github.com/containrrr/watchtower/pkg/types"
	log "github.com/sirupsen/logrus"
)

type emailNotifier struct {
	from, to               string
	server, user, password string
	port                   int
	tlsSkipVerify          bool
	entries                []*log.Entry
	delay                  time.Duration
}

func newEmailNotifier(c *cobra.Command) t.Notifier {
	flags := c.Flags()

	from, _ := flags.GetString("notification-email-from")
	to, _ := flags.GetString("notification-email-to")
	server, _ := flags.GetString("notification-email-server")
	user, _ := flags.GetString("notification-email-server-user")
	password, _ := flags.GetString("notification-email-server-password")
	port, _ := flags.GetInt("notification-email-server-port")
	tlsSkipVerify, _ := flags.GetBool("notification-email-server-tls-skip-verify")
	delay, _ := flags.GetInt("notification-email-delay")

	// Check for required configuration
	if from == "" || to == "" {
		log.Warn("Email notifier requires 'from' and 'to' addresses, skipping email notification")
		return nil
	}

	n := &emailNotifier{
		entries:       []*log.Entry{},
		from:          from,
		to:            to,
		server:        server,
		user:          user,
		password:      password,
		port:          port,
		tlsSkipVerify: tlsSkipVerify,
		delay:         time.Duration(delay) * time.Second,
	}

	return n
}

func (e *emailNotifier) StartNotification() {
	e.entries = make([]*log.Entry, 0)
}

func (e *emailNotifier) SendNotification(report t.Report) {
	if len(e.entries) == 0 && report != nil {
		log.Debug("No log entries to send via email")
		return
	}

	// Build email body
	var body strings.Builder
	for _, entry := range e.entries {
		body.WriteString(entry.Message)
		body.WriteString("\n")
	}

	// Send email
	if err := e.sendEmail(body.String()); err != nil {
		log.WithError(err).Error("Failed to send email notification")
	}
}

func (e *emailNotifier) sendEmail(body string) error {
	if e.delay > 0 {
		time.Sleep(e.delay)
	}

	// Build email address
	addr := fmt.Sprintf("%s:%d", e.server, e.port)

	// Build email message
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Watchtower Update\r\n\r\n%s",
		e.from, e.to, body)

	// Connect to SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS if supported
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: e.tlsSkipVerify,
			ServerName:         e.server,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate if username and password are provided
	if e.user != "" && e.password != "" {
		auth := smtp.PlainAuth("", e.user, e.password, e.server)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Send email
	if err := client.Mail(e.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err := client.Rcpt(e.to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}
	defer wc.Close()

	_, err = wc.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (e *emailNotifier) Close() {
	e.entries = nil
}

func (e *emailNotifier) AddLogHook() {
	hook := &logHook{
		notifier: e,
	}
	log.AddHook(hook)
}

func (e *emailNotifier) GetNames() []string {
	return []string{emailType}
}

func (e *emailNotifier) GetURLs() []string {
	return []string{}
}

// logHook is a logrus hook that captures log entries for email notifications
type logHook struct {
	notifier *emailNotifier
}

func (h *logHook) Levels() []log.Level {
	return log.AllLevels[:log.Level(log.InfoLevel)+1]
}

func (h *logHook) Fire(entry *log.Entry) error {
	if h.notifier != nil {
		h.notifier.entries = append(h.notifier.entries, entry)
	}
	return nil
}