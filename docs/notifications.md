# Notifications

Watchtower can send notifications when containers are updated via Email.

!!! note "Using multiple notifications with environment variables"
    There is currently a bug in Viper (https://github.com/spf13/viper/issues/380), which prevents comma-separated slices to
    be used when using the environment variable.
    A workaround is available where we instead put quotes around the environment variable value and replace the commas with
    spaces:
    ```
    WATCHTOWER_NOTIFICATIONS="email"
    ```
    If you're a `docker-compose` user, make sure to specify environment variables' values in your `.yml` file without double
    quotes (`"`). This prevents unexpected errors when watchtower starts.

## Settings

-   `--notifications-level` (env. `WATCHTOWER_NOTIFICATIONS_LEVEL`): Controls the log level which is used for the notifications. If omitted, the default log level is `info`. Possible values are: `panic`, `fatal`, `error`, `warn`, `info`, `debug` or `trace`.
-   `--notifications-hostname` (env. `WATCHTOWER_NOTIFICATIONS_HOSTNAME`): Custom hostname specified in subject/title. Useful to override the operating system hostname.
-   `--notifications-delay` (env. `WATCHTOWER_NOTIFICATIONS_DELAY`): Delay before sending notifications expressed in seconds.
-   Watchtower will post a notification every time it is started. This behavior [can be changed](https://containrrr.github.io/watchtower/arguments/#without_sending_a_startup_message) with an argument.
-   `--notification-title-tag` (env. `WATCHTOWER_NOTIFICATION_TITLE_TAG`): Prefix to include in the title. Useful when running multiple watchtowers.
-   `--notification-skip-title` (env. `WATCHTOWER_NOTIFICATION_SKIP_TITLE`): Do not pass the title param to notifications.

## Email Notifications

To send notifications via Email, the following command-line options, or their corresponding environment variables, can be set:

-   `--notification-email-from` (env. `WATCHTOWER_NOTIFICATION_EMAIL_FROM`): The sender email address.
-   `--notification-email-to` (env. `WATCHTOWER_NOTIFICATION_EMAIL_TO`): The recipient email address.
-   `--notification-email-server` (env. `WATCHTOWER_NOTIFICATION_EMAIL_SERVER`): The SMTP server address.
-   `--notification-email-server-port` (env. `WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PORT`): The SMTP server port (default: 25).
-   `--notification-email-server-user` (env. `WATCHTOWER_NOTIFICATION_EMAIL_SERVER_USER`): The SMTP server username for authentication.
-   `--notification-email-server-password` (env. `WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PASSWORD`): The SMTP server password for authentication. This option can also reference a file, in which case the contents of the file are used.
-   `--notification-email-server-tls-skip-verify` (env. `WATCHTOWER_NOTIFICATION_EMAIL_SERVER_TLS_SKIP_VERIFY`): Controls whether watchtower verifies the SMTP server's certificate chain and host name. Should only be used for testing.
-   `--notification-email-delay` (env. `WATCHTOWER_NOTIFICATION_EMAIL_DELAY`): Delay before sending email notifications expressed in seconds.
-   `--notification-email-subjecttag` (env. `WATCHTOWER_NOTIFICATION_EMAIL_SUBJECTTAG`): Subject prefix tag for notifications via mail.

### Example docker-compose usage

```yaml
services:
  watchtower:
    image: containrrr/watchtower
    environment:
      - WATCHTOWER_NOTIFICATIONS=email
      - WATCHTOWER_NOTIFICATION_EMAIL_FROM=watchtower@example.com
      - WATCHTOWER_NOTIFICATION_EMAIL_TO=admin@example.com
      - WATCHTOWER_NOTIFICATION_EMAIL_SERVER=smtp.example.com
      - WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PORT=587
      - WATCHTOWER_NOTIFICATION_EMAIL_SERVER_USER=user@example.com
      - WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PASSWORD=password
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

### Example command-line usage

```bash
docker run -d \
  --name watchtower \
  -e WATCHTOWER_NOTIFICATIONS=email \
  -e WATCHTOWER_NOTIFICATION_EMAIL_FROM=watchtower@example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_TO=admin@example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER=smtp.example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PORT=587 \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER_USER=user@example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PASSWORD=password \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower
```

### SMTP Authentication

If your SMTP server requires authentication, you need to provide both the username and password:

```yaml
environment:
  - WATCHTOWER_NOTIFICATION_EMAIL_SERVER_USER=your-username
  - WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PASSWORD=your-password
```

### TLS Configuration

By default, watchtower will attempt to use STARTTLS if available. If you need to disable TLS verification (not recommended for production):

```yaml
environment:
  - WATCHTOWER_NOTIFICATION_EMAIL_SERVER_TLS_SKIP_VERIFY=true
```

### Multiple Recipients

Currently, watchtower only supports a single recipient per configuration. If you need to send notifications to multiple recipients, you can:

1. Configure your SMTP server to forward messages to multiple recipients
2. Use multiple watchtower instances with different configurations
3. Use an email distribution list or mailing list

## Testing Notifications

To test your notification configuration, you can run watchtower once and check if you receive an email:

```bash
docker run --rm \
  -e WATCHTOWER_NOTIFICATIONS=email \
  -e WATCHTOWER_NOTIFICATION_EMAIL_FROM=watchtower@example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_TO=admin@example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER=smtp.example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PORT=587 \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER_USER=user@example.com \
  -e WATCHTOWER_NOTIFICATION_EMAIL_SERVER_PASSWORD=password \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower --run-once
```

## Troubleshooting

### No Email Received

1. Check your SMTP server configuration
2. Verify your authentication credentials
3. Check watchtower logs for errors: `docker logs watchtower`
4. Ensure your firewall allows outbound connections to the SMTP server
5. Check if your email provider requires specific port or security settings

### Authentication Errors

1. Verify your username and password are correct
2. Check if your email provider requires app-specific passwords
3. Ensure your account allows third-party applications

### TLS Errors

1. Check if your SMTP server uses a valid SSL certificate
2. If using a self-signed certificate, consider using `--notification-email-server-tls-skip-verify=true` (not recommended for production)
3. Verify the server hostname matches the certificate