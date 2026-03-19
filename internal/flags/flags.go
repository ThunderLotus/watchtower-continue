package flags

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/containrrr/watchtower/pkg/retry"
)

// DockerAPIMinVersion is the minimum version of the docker api required to
// use watchtower
const DockerAPIMinVersion string = "1.25"

var defaultInterval = int((time.Hour * 24).Seconds())

// RegisterDockerFlags that are used directly by the docker api client
func RegisterDockerFlags(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()
	flags.StringP("host", "H", envString("DOCKER_HOST"), "daemon socket to connect to")
	flags.BoolP("tlsverify", "v", envBool("DOCKER_TLS_VERIFY"), "use TLS and verify the remote")
	flags.StringP("api-version", "a", envString("DOCKER_API_VERSION"), "api version to use by docker client")
}

// RegisterSystemFlags that are used by watchtower to modify the program flow
func RegisterSystemFlags(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()
	flags.IntP(
		"interval",
		"i",
		envInt("WATCHTOWER_POLL_INTERVAL"),
		"Poll interval (in seconds)")

	flags.StringP(
		"schedule",
		"s",
		envString("WATCHTOWER_SCHEDULE"),
		"The cron expression which defines when to update")

	flags.DurationP(
		"stop-timeout",
		"t",
		envDuration("WATCHTOWER_TIMEOUT"),
		"Timeout before a container is forcefully stopped")

	flags.BoolP(
		"no-pull",
		"",
		envBool("WATCHTOWER_NO_PULL"),
		"Do not pull any new images")

	flags.BoolP(
		"no-restart",
		"",
		envBool("WATCHTOWER_NO_RESTART"),
		"Do not restart any containers")

	flags.BoolP(
		"no-startup-message",
		"",
		envBool("WATCHTOWER_NO_STARTUP_MESSAGE"),
		"Prevents watchtower from sending a startup message")

	flags.BoolP(
		"cleanup",
		"c",
		envBool("WATCHTOWER_CLEANUP"),
		"Remove previously used images after updating")

	flags.BoolP(
		"remove-volumes",
		"",
		envBool("WATCHTOWER_REMOVE_VOLUMES"),
		"Remove attached volumes before updating")

	flags.BoolP(
		"label-enable",
		"e",
		envBool("WATCHTOWER_LABEL_ENABLE"),
		"Watch containers where the com.centurylinklabs.watchtower.enable label is true")

	flags.StringSliceP(
		"disable-containers",
		"x",
		// Due to issue spf13/viper#380, can't use viper.GetStringSlice:
		regexp.MustCompile("[, ]+").Split(envString("WATCHTOWER_DISABLE_CONTAINERS"), -1),
		"Comma-separated list of containers to explicitly exclude from watching.")

	flags.StringP(
		"log-format",
		"l",
		viper.GetString("WATCHTOWER_LOG_FORMAT"),
		"Sets what logging format to use for console output. Possible values: Auto, LogFmt, Pretty, JSON")

	flags.BoolP(
		"debug",
		"d",
		envBool("WATCHTOWER_DEBUG"),
		"Enable debug mode with verbose logging")

	flags.BoolP(
		"trace",
		"",
		envBool("WATCHTOWER_TRACE"),
		"Enable trace mode with very verbose logging - caution, exposes credentials")

	flags.BoolP(
		"monitor-only",
		"m",
		envBool("WATCHTOWER_MONITOR_ONLY"),
		"Will only monitor for new images, not update the containers")

	flags.BoolP(
		"run-once",
		"R",
		envBool("WATCHTOWER_RUN_ONCE"),
		"Run once now and exit")

	flags.BoolP(
		"include-restarting",
		"",
		envBool("WATCHTOWER_INCLUDE_RESTARTING"),
		"Will also include restarting containers")

	flags.BoolP(
		"include-stopped",
		"S",
		envBool("WATCHTOWER_INCLUDE_STOPPED"),
		"Will also include created and exited containers")

	flags.BoolP(
		"revive-stopped",
		"",
		envBool("WATCHTOWER_REVIVE_STOPPED"),
		"Will also start stopped containers that were updated, if include-stopped is active")

	flags.BoolP(
		"enable-lifecycle-hooks",
		"",
		envBool("WATCHTOWER_LIFECYCLE_HOOKS"),
		"Enable the execution of commands triggered by pre- and post-update lifecycle hooks")

	flags.BoolP(
		"rolling-restart",
		"",
		envBool("WATCHTOWER_ROLLING_RESTART"),
		"Restart containers one at a time")

	// https://no-color.org/
	flags.BoolP(
		"no-color",
		"",
		viper.IsSet("NO_COLOR"),
		"Disable ANSI color escape codes in log output")

	flags.StringP(
		"scope",
		"",
		envString("WATCHTOWER_SCOPE"),
		"Defines a monitoring scope for the Watchtower instance.")

	flags.StringP(
		"porcelain",
		"P",
		envString("WATCHTOWER_PORCELAIN"),
		`Write session results to stdout using a stable versioned format. Supported values: "v1"`)

	flags.String(
		"log-level",
		envString("WATCHTOWER_LOG_LEVEL"),
		"The maximum log level that will be written to STDERR. Possible values: panic, fatal, error, warn, info or debug")

	flags.BoolP(
		"health-check",
		"",
		false,
		"Do health check and exit")

	flags.BoolP(
		"label-take-precedence",
		"",
		envBool("WATCHTOWER_LABEL_TAKE_PRECEDENCE"),
		"Label applied to containers take precedence over arguments")

	// Retry configuration flags
	flags.BoolP(
		"retry-enable",
		"",
		envBool("WATCHTOWER_RETRY_ENABLE"),
		"Enable retry mechanism for network operations (default: true)")

	flags.IntP(
		"retry-max-attempts",
		"",
		envInt("WATCHTOWER_RETRY_MAX_ATTEMPTS"),
		"Maximum number of retry attempts for network operations (default: 5)")

	flags.DurationP(
		"retry-initial-delay",
		"",
		envDuration("WATCHTOWER_RETRY_INITIAL_DELAY"),
		"Initial delay before first retry (default: 1s)")

	flags.DurationP(
		"retry-max-delay",
		"",
		envDuration("WATCHTOWER_RETRY_MAX_DELAY"),
		"Maximum delay between retries (default: 30s)")
}

// RegisterNotificationFlags that are used by watchtower to send notifications
func RegisterNotificationFlags(rootCmd *cobra.Command) {
	flags := rootCmd.PersistentFlags()

	flags.StringSliceP(
		"notifications",
		"n",
		envStringSlice("WATCHTOWER_NOTIFICATIONS"),
		"Notification types to send (valid: email)")

	flags.String(
		"notifications-level",
		envString("WATCHTOWER_NOTIFICATIONS_LEVEL"),
		"The log level used for sending notifications. Possible values: panic, fatal, error, warn, info or debug")

	flags.IntP(
		"notifications-delay",
		"",
		envInt("WATCHTOWER_NOTIFICATIONS_DELAY"),
		"Delay before sending notifications, expressed in seconds")

	flags.StringP(
		"notifications-hostname",
		"",
		envString("WATCHTOWER_NOTIFICATIONS_HOSTNAME"),
		"Custom hostname for notification titles")

	flags.StringP(
		"notification-email-from",
		"",
		envString("WATCHTOWER_NOTIFICATION_EMAIL_FROM"),
		"Address to send notification emails from")

	flags.StringP(
		"notification-email-to",
		"",
		envString("WATCHTOWER_NOTIFICATION_EMAIL_TO"),
		"Address to send notification emails to")

	flags.IntP(
		"notification-email-delay",
		"",
		envInt("WATCHTOWER_NOTIFICATION_EMAIL_DELAY"),
		"Delay before sending notifications, expressed in seconds")

	flags.StringP(
		"notification-email-server",
		"",
		envString("WATCHTOWER_NOTIFICATION_EMAIL_SERVER"),
		"SMTP server to send notification emails through")

	flags.IntP(
		"notification-email-server-port",
		"",
		envInt("WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PORT"),
		"SMTP server port to send notification emails through")

	flags.BoolP(
		"notification-email-server-tls-skip-verify",
		"",
		envBool("WATCHTOWER_NOTIFICATION_EMAIL_SERVER_TLS_SKIP_VERIFY"),
		`Controls whether watchtower verifies the SMTP server's certificate chain and host name.
Should only be used for testing.`)

	flags.StringP(
		"notification-email-server-user",
		"",
		envString("WATCHTOWER_NOTIFICATION_EMAIL_SERVER_USER"),
		"SMTP server user for sending notifications")

	flags.StringP(
		"notification-email-server-password",
		"",
		envString("WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PASSWORD"),
		"SMTP server password for sending notifications")

	flags.StringP(
		"notification-email-subjecttag",
		"",
		envString("WATCHTOWER_NOTIFICATION_EMAIL_SUBJECTTAG"),
		"Subject prefix tag for notifications via mail")

	flags.StringP(
		"notification-title-tag",
		"",
		envString("WATCHTOWER_NOTIFICATION_TITLE_TAG"),
		"Title prefix tag for notifications")

	flags.Bool("notification-skip-title",
		envBool("WATCHTOWER_NOTIFICATION_SKIP_TITLE"),
		"Do not pass the title param to notifications")

	flags.String(
		"warn-on-head-failure",
		envString("WATCHTOWER_WARN_ON_HEAD_FAILURE"),
		"When to warn about HEAD pull requests failing. Possible values: always, auto or never")
}

func envString(key string) string {
	viper.MustBindEnv(key)
	return viper.GetString(key)
}

func envStringSlice(key string) []string {
	viper.MustBindEnv(key)
	return viper.GetStringSlice(key)
}

func envInt(key string) int {
	viper.MustBindEnv(key)
	return viper.GetInt(key)
}

func envBool(key string) bool {
	viper.MustBindEnv(key)
	return viper.GetBool(key)
}

func envFloat64(key string) float64 {
	viper.MustBindEnv(key)
	return viper.GetFloat64(key)
}

func envDuration(key string) time.Duration {
	viper.MustBindEnv(key)
	return viper.GetDuration(key)
}

// SetDefaults provides default values for environment variables
func SetDefaults() {
	viper.AutomaticEnv()
	viper.SetDefault("DOCKER_HOST", "unix:///var/run/docker.sock")
	// Removed default DOCKER_API_VERSION to allow auto-negotiation
	// This ensures compatibility with both old and new Docker daemon versions
	viper.SetDefault("WATCHTOWER_POLL_INTERVAL", defaultInterval)
	viper.SetDefault("WATCHTOWER_TIMEOUT", time.Second*10)
	viper.SetDefault("WATCHTOWER_NOTIFICATIONS", []string{})
	viper.SetDefault("WATCHTOWER_NOTIFICATIONS_LEVEL", "info")
	viper.SetDefault("WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PORT", 25)
	viper.SetDefault("WATCHTOWER_NOTIFICATION_EMAIL_SUBJECTTAG", "")
	viper.SetDefault("WATCHTOWER_LOG_LEVEL", "info")
	viper.SetDefault("WATCHTOWER_LOG_FORMAT", "auto")
	viper.SetDefault("WATCHTOWER_RETRY_ENABLE", true)
	viper.SetDefault("WATCHTOWER_RETRY_MAX_ATTEMPTS", 5)
	viper.SetDefault("WATCHTOWER_RETRY_INITIAL_DELAY", 1*time.Second)
	viper.SetDefault("WATCHTOWER_RETRY_MAX_DELAY", 30*time.Second)
}

// EnvConfig translates the command-line options into environment variables
// that will initialize the api client
func EnvConfig(cmd *cobra.Command) error {
	var err error
	var host string
	var tls bool
	var version string

	flags := cmd.PersistentFlags()

	if host, err = flags.GetString("host"); err != nil {
		return err
	}
	if tls, err = flags.GetBool("tlsverify"); err != nil {
		return err
	}
	if version, err = flags.GetString("api-version"); err != nil {
		return err
	}
	if err = setEnvOptStr("DOCKER_HOST", host); err != nil {
		return err
	}
	if err = setEnvOptBool("DOCKER_TLS_VERIFY", tls); err != nil {
		return err
	}
	if err = setEnvOptStr("DOCKER_API_VERSION", version); err != nil {
		return err
	}
	return nil
}

// ReadFlags reads common flags used in the main program flow of watchtower
func ReadFlags(cmd *cobra.Command) (bool, bool, bool, time.Duration, error) {
	flags := cmd.PersistentFlags()

	var err error
	var cleanup bool
	var noRestart bool
	var monitorOnly bool
	var timeout time.Duration

	if cleanup, err = flags.GetBool("cleanup"); err != nil {
		return false, false, false, 0, fmt.Errorf("failed to get cleanup flag: %w", err)
	}
	if noRestart, err = flags.GetBool("no-restart"); err != nil {
		return false, false, false, 0, fmt.Errorf("failed to get no-restart flag: %w", err)
	}
	if monitorOnly, err = flags.GetBool("monitor-only"); err != nil {
		return false, false, false, 0, fmt.Errorf("failed to get monitor-only flag: %w", err)
	}
	if timeout, err = flags.GetDuration("stop-timeout"); err != nil {
		return false, false, false, 0, fmt.Errorf("failed to get stop-timeout flag: %w", err)
	}

	return cleanup, noRestart, monitorOnly, timeout, nil
}

func setEnvOptStr(env string, opt string) error {
	if opt == "" || opt == os.Getenv(env) {
		return nil
	}
	err := os.Setenv(env, opt)
	if err != nil {
		return err
	}
	return nil
}

func setEnvOptBool(env string, opt bool) error {
	if opt {
		return setEnvOptStr(env, "1")
	}
	return nil
}

// GetSecretsFromFiles checks if passwords/tokens/webhooks have been passed as a file instead of plaintext.
// If so, the value of the flag will be replaced with the contents of the file.
func GetSecretsFromFiles(rootCmd *cobra.Command) error {
	flags := rootCmd.PersistentFlags()

	secrets := []string{
		"notification-email-server-password",
	}
	for _, secret := range secrets {
		if err := getSecretFromFile(flags, secret); err != nil {
			return fmt.Errorf("failed to get secret from flag %v: %w", secret, err)
		}
	}
	return nil
}

// getSecretFromFile will check if the flag contains a reference to a file; if it does, replaces the value of the flag with the contents of the file.
func getSecretFromFile(flags *pflag.FlagSet, secret string) error {
	flag := flags.Lookup(secret)
	if sliceValue, ok := flag.Value.(pflag.SliceValue); ok {
		oldValues := sliceValue.GetSlice()
		values := make([]string, 0, len(oldValues))
		for _, value := range oldValues {
			if value != "" && isFile(value) {
				file, err := os.Open(value)
				if err != nil {
					return err
				}
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					if line == "" {
						continue
					}
					values = append(values, line)
				}
				if err := file.Close(); err != nil {
					return err
				}
			} else {
				values = append(values, value)
			}
		}
		return sliceValue.Replace(values)
	}

	value := flag.Value.String()
	if value != "" && isFile(value) {
		content, err := os.ReadFile(value)
		if err != nil {
			return err
		}
		return flags.Set(secret, strings.TrimSpace(string(content)))
	}

	return nil
}

func isFile(s string) bool {
	firstColon := strings.IndexRune(s, ':')
	if firstColon != 1 && firstColon != -1 {
		// If the string contains a ':', but it's not the second character, it's probably not a file
		// and will cause a fatal error on windows if stat'ed
		// This still allows for paths that start with 'c:\' etc.
		return false
	}
	_, err := os.Stat(s)
	return !errors.Is(err, os.ErrNotExist)
}

// ProcessFlagAliases updates the value of flags that are being set by helper flags
func ProcessFlagAliases(flags *pflag.FlagSet) error {

	porcelain, err := flags.GetString(`porcelain`)
	if err != nil {
		return fmt.Errorf("failed to get porcelain flag: %w", err)
	}
	if porcelain != "" {
		if porcelain != "v1" {
			return fmt.Errorf("unknown porcelain version %q. Supported values: \"v1\"", porcelain)
		}
		// Porcelain mode is not supported without Shoutrrr
		log.Warn("Porcelain mode is not supported in the simplified notification system")
	}

	scheduleChanged := flags.Changed(`schedule`)
	intervalChanged := flags.Changed(`interval`)
	// FIXME: snakeswap
	// due to how viper is integrated by swapping the defaults for the flags, we need this hack:
	if val, _ := flags.GetString(`schedule`); val != `` {
		scheduleChanged = true
	}
	if val, _ := flags.GetInt(`interval`); val != defaultInterval {
		intervalChanged = true
	}

	if intervalChanged && scheduleChanged {
		return errors.New("only schedule or interval can be defined, not both")
	}

	// update schedule flag to match interval if it's set, or to the default if none of them are
	if intervalChanged || !scheduleChanged {
		interval, _ := flags.GetInt(`interval`)
		_ = flags.Set(`schedule`, fmt.Sprintf(`@every %ds`, interval))
	}

	if flagIsEnabled(flags, `debug`) {
		_ = flags.Set(`log-level`, `debug`)
	}

	if flagIsEnabled(flags, `trace`) {
		_ = flags.Set(`log-level`, `trace`)
	}

	return nil
}


// SetupLogging reads only the flags that is needed to set up logging and applies them to the global logger
func SetupLogging(f *pflag.FlagSet) error {
	logFormat, _ := f.GetString(`log-format`)
	noColor, _ := f.GetBool("no-color")

	// Configure formatter
	switch strings.ToLower(logFormat) {
	case "auto":
		// This will either use the "pretty" or "logfmt" format, based on whether the standard out is connected to a TTY
		log.SetFormatter(&log.TextFormatter{
			DisableColors: noColor,
			// enable logrus built-in support for https://bixense.com/clicolors/
			EnvironmentOverrideColors: true,
			FullTimestamp:             true,
			TimestampFormat:           time.RFC3339,
		})
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case "logfmt":
		log.SetFormatter(&log.TextFormatter{
			DisableColors:   true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	case "pretty":
		log.SetFormatter(&log.TextFormatter{
			// "Pretty" format combined with `--no-color` will only change the timestamp to the time since start
			ForceColors:      !noColor,
			FullTimestamp:    false,
			EnvironmentOverrideColors: true,
		})
	default:
		return fmt.Errorf("invalid log format: %s", logFormat)
	}

	// Configure log level
	rawLogLevel, _ := f.GetString(`log-level`)
	if logLevel, err := log.ParseLevel(rawLogLevel); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	} else {
		log.SetLevel(logLevel)
	}

	// Always output to stderr
	log.SetOutput(os.Stderr)

	return nil
}

func flagIsEnabled(flags *pflag.FlagSet, name string) bool {
	value, err := flags.GetBool(name)
	if err != nil {
		log.Warnf(`The flag %q is not defined: %v`, name, err)
		return false
	}
	return value
}

func appendFlagValue(flags *pflag.FlagSet, name string, values ...string) error {
	flag := flags.Lookup(name)
	if flag == nil {
		return fmt.Errorf(`invalid flag name %q`, name)
	}

	if flagValues, ok := flag.Value.(pflag.SliceValue); ok {
		for _, value := range values {
			_ = flagValues.Append(value)
		}
	} else {
		return fmt.Errorf(`the value for flag %q is not a slice value`, name)
	}

	return nil
}

func setFlagIfDefault(flags *pflag.FlagSet, name string, value string) {
	if flags.Changed(name) {
		return
	}
	if err := flags.Set(name, value); err != nil {
		log.Errorf(`Failed to set flag: %v`, err)
	}
}

// GetRetryConfig returns the retry configuration from flags
func GetRetryConfig(cmd *cobra.Command) (*retry.Config, error) {
	flags := cmd.PersistentFlags()

	enable, err := flags.GetBool("retry-enable")
	if err != nil {
		return nil, fmt.Errorf("failed to get retry-enable flag: %w", err)
	}

	maxAttempts, err := flags.GetInt("retry-max-attempts")
	if err != nil {
		return nil, fmt.Errorf("failed to get retry-max-attempts flag: %w", err)
	}

	initialDelay, err := flags.GetDuration("retry-initial-delay")
	if err != nil {
		return nil, fmt.Errorf("failed to get retry-initial-delay flag: %w", err)
	}

	maxDelay, err := flags.GetDuration("retry-max-delay")
	if err != nil {
		return nil, fmt.Errorf("failed to get retry-max-delay flag: %w", err)
	}

	// Validate retry configuration
	if maxAttempts < 1 {
		return nil, fmt.Errorf("retry-max-attempts must be at least 1, got %d", maxAttempts)
	}
	if maxAttempts > 100 {
		log.Warnf("retry-max-attempts is very high (%d), consider reducing it", maxAttempts)
	}

	if initialDelay < 0 {
		return nil, fmt.Errorf("retry-initial-delay must be non-negative, got %v", initialDelay)
	}
	if initialDelay > 5*time.Minute {
		log.Warnf("retry-initial-delay is very long (%v), consider reducing it", initialDelay)
	}

	if maxDelay < 0 {
		return nil, fmt.Errorf("retry-max-delay must be non-negative, got %v", maxDelay)
	}
	if maxDelay < initialDelay {
		return nil, fmt.Errorf("retry-max-delay (%v) must be greater than or equal to retry-initial-delay (%v)", maxDelay, initialDelay)
	}
	if maxDelay > 1*time.Hour {
		log.Warnf("retry-max-delay is very long (%v), consider reducing it", maxDelay)
	}

	config := retry.DefaultConfig()
	config.EnableRetry = enable
	config.MaxAttempts = uint64(maxAttempts)
	config.InitialDelay = initialDelay
	config.MaxDelay = maxDelay

	return config, nil
}
