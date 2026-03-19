package notifications

import (
	"os"
	"strings"

	ty "github.com/containrrr/watchtower/pkg/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	emailType = "email"
)

// NewNotifier creates and returns a new Notifier, using global configuration.
func NewNotifier(c *cobra.Command) ty.Notifier {
	notifyEntry := log.WithFields(log.Fields{
		"module":    "notifications",
		"operation": "initialize_notifier",
	})

	f := c.Flags()

	level, _ := f.GetString("notifications-level")
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		notifyEntry.WithError(err).WithField("provided_level", level).
			Warn("Invalid notifications log level, using default 'info'")
		logLevel = log.InfoLevel
	}

	reportTemplate, _ := f.GetBool("notification-report")
	stdout, _ := f.GetBool("notification-log-stdout")

	notifyEntry.WithFields(log.Fields{
		"log_level":       logLevel.String(),
		"report_template": reportTemplate,
		"stdout":          stdout,
	}).Debug("Notifier configuration loaded")

	// Parse types and create notifiers
	types, err := c.Flags().GetStringSlice("notifications")
	if err != nil {
		notifyEntry.WithError(err).Warn("Could not read notifications argument, notifications will be disabled")
		return &dummyNotifier{}
	}

	notifyEntry.WithField("notification_types", types).Debug("Processing notification types")

	var notifier ty.Notifier
	enabledTypes := make([]string, 0)

	for _, t := range types {
		switch t {
		case emailType:
			notifier = newEmailNotifier(c)
			enabledTypes = append(enabledTypes, emailType)
		default:
			notifyEntry.WithField("type", t).Warnf("Unknown notification type '%s', only 'email' is supported", t)
		}
	}

	if notifier == nil {
		notifyEntry.Info("No notifiers configured, notifications will be disabled")
		return &dummyNotifier{}
	}

	notifyEntry.WithField("enabled_types", enabledTypes).Info("Notifier created successfully")
	return notifier
}

// GetTitle formats the title based on the passed hostname and tag
func GetTitle(hostname string, tag string) string {
	tb := strings.Builder{}

	if tag != "" {
		tb.WriteRune('[')
		tb.WriteString(tag)
		tb.WriteRune(']')
		tb.WriteRune(' ')
	}

	tb.WriteString("Watchtower updates")

	if hostname != "" {
		tb.WriteString(" on ")
		tb.WriteString(hostname)
	}

	return tb.String()
}

// GetTemplateData populates the static notification data from flags and environment
func GetTemplateData(c *cobra.Command) StaticData {
	f := c.PersistentFlags()

	hostname, _ := f.GetString("notifications-hostname")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	title := ""
	if skip, _ := f.GetBool("notification-skip-title"); !skip {
		tag, _ := f.GetString("notification-title-tag")
		if tag == "" {
			// For legacy email support
			tag, _ = f.GetString("notification-email-subjecttag")
		}
		title = GetTitle(hostname, tag)
	}

	return StaticData{
		Host:  hostname,
		Title: title,
	}
}

// ColorHex is the default notification color used for services that support it (formatted as a CSS hex string)
const ColorHex = "#406170"

// dummyNotifier is a no-op notifier used when notification initialization fails
type dummyNotifier struct{}

func (d *dummyNotifier) StartNotification() {}
func (d *dummyNotifier) SendNotification(c ty.Report) {}
func (d *dummyNotifier) Close() {}
func (d *dummyNotifier) AddLogHook() {}
func (d *dummyNotifier) GetNames() []string { return []string{} }
func (d *dummyNotifier) GetURLs() []string { return []string{} }

// ColorInt is the default notification color used for services that support it (as an int value)
const ColorInt = 0x406170