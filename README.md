# Grafana On-call Notifier

This is a simple On-Call reminder application. It uses Grafana On-Call API to get a list of every schedule and send reminders for them.

## Config

You can edit configurations in `config.yml` file.

```yml
log_level: info # info, debug

email_server:
  host: mail.emailserver.com
  port: 465
  username: username@mail.emailserver.com
  password: "PASSWORD"

grafana_oncall:
  url: URL_OF_ONCALL_APPLICATION
  admin_token: oncall-admin-token

```

### Output Email

```text
Dear oncall-user,
Your next on-call shift in schedule: Iaas-SRE-Backup is about to begin shortly. Here's a summary of that shift:
2025-05-16 00:00:00 +0330 +0330 - 2025-05-17 00:00:00 +0330 +0330: oncall-user@domain.com (You)

Best Regards,
Oncall-Notifier
```
