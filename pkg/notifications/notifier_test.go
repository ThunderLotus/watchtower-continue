package notifications_test

import (
	"os"

	"github.com/containrrr/watchtower/cmd"
	"github.com/containrrr/watchtower/internal/flags"
	"github.com/containrrr/watchtower/pkg/notifications"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("notifications", func() {
	Describe("the notifier", func() {
		When("only empty notifier types are provided", func() {
			command := cmd.NewRootCommand()
			flags.RegisterNotificationFlags(command)

			err := command.ParseFlags([]string{})
			Expect(err).NotTo(HaveOccurred())
			notif := notifications.NewNotifier(command)

			Expect(notif.GetNames()).To(BeEmpty())
		})

		When("unknown notifier type is provided", func() {
			command := cmd.NewRootCommand()
			flags.RegisterNotificationFlags(command)

			err := command.ParseFlags([]string{
				"--notifications",
				"unknown-type",
			})
			Expect(err).NotTo(HaveOccurred())
			notif := notifications.NewNotifier(command)

			Expect(notif.GetNames()).To(BeEmpty())
		})

		When("title is overridden in flag", func() {
			It("should use the specified hostname in the title", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				err := command.ParseFlags([]string{
					"--notifications-hostname",
					"test.host",
				})
				Expect(err).NotTo(HaveOccurred())
				data := notifications.GetTemplateData(command)
				title := data.Title
				Expect(title).To(Equal("Watchtower updates on test.host"))
			})
		})

		When("no hostname can be resolved", func() {
			It("should use the default simple title", func() {
				title := notifications.GetTitle("", "")
				Expect(title).To(Equal("Watchtower updates"))
			})
		})

		When("title tag is set", func() {
			It("should use the prefix in the title", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				Expect(command.ParseFlags([]string{
					"--notification-title-tag",
					"PREFIX",
				})).To(Succeed())

				data := notifications.GetTemplateData(command)
				Expect(data.Title).To(HavePrefix("[PREFIX]"))
			})
		})

		When("legacy email tag is set", func() {
			It("should use the prefix in the title", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				Expect(command.ParseFlags([]string{
					"--notification-email-subjecttag",
					"PREFIX",
				})).To(Succeed())

				data := notifications.GetTemplateData(command)
				Expect(data.Title).To(HavePrefix("[PREFIX]"))
			})
		})

		When("the skip title flag is set", func() {
			It("should return an empty title", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				Expect(command.ParseFlags([]string{
					"--notification-skip-title",
				})).To(Succeed())

				data := notifications.GetTemplateData(command)
				Expect(data.Title).To(BeEmpty())
			})
		})

		When("hostname is set and title tag is provided", func() {
			It("should combine both in the title", func() {
				title := notifications.GetTitle("myhost", "PROD")
				Expect(title).To(Equal("[PROD] Watchtower updates on myhost"))
			})
		})
	})

	Describe("the email notifier", func() {
		When("email notifier is configured", func() {
			It("should create an email notifier successfully", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				args := []string{
					"--notifications",
					"email",
					"--notification-email-from",
					"sender@example.com",
					"--notification-email-to",
					"receiver@example.com",
					"--notification-email-server",
					"mail.example.com",
					"--notification-email-server-user",
					"user",
					"--notification-email-server-password",
					"password",
				}

				err := command.ParseFlags(args)
				Expect(err).NotTo(HaveOccurred())

				notif := notifications.NewNotifier(command)
				Expect(notif).NotTo(BeNil())
				Expect(notif.GetNames()).To(ContainElement("email"))
			})
		})

		When("email notifier is configured with all options", func() {
			It("should create an email notifier with all options", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				args := []string{
					"--notifications",
					"email",
					"--notification-email-from",
					"sender@example.com",
					"--notification-email-to",
					"receiver@example.com",
					"--notification-email-server",
					"mail.example.com",
					"--notification-email-server-port",
					"587",
					"--notification-email-server-user",
					"user",
					"--notification-email-server-password",
					"password",
					"--notification-email-server-tls-skip-verify",
					"--notification-email-delay",
					"10",
				}

				err := command.ParseFlags(args)
				Expect(err).NotTo(HaveOccurred())

				notif := notifications.NewNotifier(command)
				Expect(notif).NotTo(BeNil())
				Expect(notif.GetNames()).To(ContainElement("email"))
			})
		})

		When("email notifier is configured with minimal options", func() {
			It("should create an email notifier with minimal options", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				args := []string{
					"--notifications",
					"email",
					"--notification-email-from",
					"sender@example.com",
					"--notification-email-to",
					"receiver@example.com",
				}

				err := command.ParseFlags(args)
				Expect(err).NotTo(HaveOccurred())

				notif := notifications.NewNotifier(command)
				Expect(notif).NotTo(BeNil())
				Expect(notif.GetNames()).To(ContainElement("email"))
			})
		})

		When("email notifier uses system hostname", func() {
			It("should get hostname from system", func() {
				hostname, err := os.Hostname()
				Expect(err).NotTo(HaveOccurred())

				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				data := notifications.GetTemplateData(command)
				Expect(data.Host).To(Equal(hostname))
			})
		})

		When("email notifier is used without configuration", func() {
			It("should not create an email notifier", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				args := []string{
					"--notifications",
					"email",
				}

				err := command.ParseFlags(args)
				Expect(err).NotTo(HaveOccurred())

				notif := notifications.NewNotifier(command)
				Expect(notif.GetNames()).To(BeEmpty())
			})
		})
	})

	Describe("notification lifecycle", func() {
		When("starting and closing notifications", func() {
			It("should handle start and close operations", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				args := []string{
					"--notifications",
					"email",
					"--notification-email-from",
					"sender@example.com",
					"--notification-email-to",
					"receiver@example.com",
					"--notification-email-server",
					"mail.example.com",
				}

				err := command.ParseFlags(args)
				Expect(err).NotTo(HaveOccurred())

				notif := notifications.NewNotifier(command)
				Expect(notif).NotTo(BeNil())

				// Start notification
				notif.StartNotification()

				// Close notification
				notif.Close()

				// Should not panic
				Expect(true).To(BeTrue())
			})
		})
	})
})