# OpenSMTPD Docker Relay

A Docker-based OpenSMTPD setup that acts as a secure email relay to Microsoft 365, with configurable access control for allowed hosts.

## Overview

**OpenSMTPD Docker Relay** is a containerized SMTP relay server using OpenSMTPD that forwards emails to Microsoft 365's SMTP servers. It includes access control to restrict relaying to only specified hosts/networks.

## Features

- 📧 **SMTP Relay** - Secure email forwarding to Microsoft 365
- 🔒 **Access Control** - Restrict relaying to allowed hosts/networks
- 🐳 **Docker Containerized** - Easy deployment and portability
- 📝 **Comprehensive Logging** - Integrated rsyslog and supervisord logging
- 🔧 **Configurable** - Easy-to-modify configuration files
- 🌐 **Multi-Port Support** - Standard SMTP (25), SMTPS (465), and Submission (587)

## Language Composition

This repository is composed of the following languages:

- **Shell** - 66.3%
- **PowerShell** - 30.5%
- **Dockerfile** - 3.2%

## Prerequisites

- Docker
- Docker Compose
- Access to Microsoft 365 SMTP servers (typically `yourdomain.mail.protection.outlook.com`)

## Quick Start

### Using Docker Compose

1. Clone this repository:
```bash
git clone <your-repo-url>
cd opensmtpd
```

2. Configure your settings:
   - Edit `allowed_hosts` to include your allowed IP addresses/networks (one per line)
   - Update `smtpd.conf` with your Microsoft 365 SMTP endpoint
   - Modify `docker-compose.yml` environment variables if needed

3. Build and run:
```bash
docker-compose up -d
```

### Using Docker CLI

```bash
docker build -t opensmtpd-relay .

docker run -d \
  --name opensmtpd-relay \
  --restart always \
  -p 25:25 \
  -p 465:465 \
  -p 587:587 \
  -e TZ=America/Chicago \
  -v $(pwd)/rsyslog.conf:/etc/rsyslog.conf:ro \
  -v $(pwd)/smtpd.conf:/etc/smtpd.conf:ro \
  -v $(pwd)/supervisord.conf:/etc/supervisor/supervisord.conf:ro \
  -v $(pwd)/allowed_hosts:/etc/mail_allowed_hosts:ro \
  opensmtpd-relay
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TZ` | Timezone for logging | America/Chicago |

### Configuration Files

#### `allowed_hosts`
List of allowed IP addresses or networks that can relay emails, one per line:
```
192.168.1.0/24
10.0.0.1
203.0.113.5
```

#### `smtpd.conf`
OpenSMTPD configuration file. Key settings:
- `listen on 0.0.0.0/0` - Listen on all interfaces
- `table allowed_hosts file:/etc/mail_allowed_hosts` - Load allowed hosts table
- `action "relay_out" relay host smtp://yourdomain.mail.protection.outlook.com` - Relay to Microsoft 365

#### `rsyslog.conf`
Logging configuration for rsyslog.

#### `supervisord.conf`
Supervisor configuration to manage OpenSMTPD and rsyslog processes.

## Usage

Once running, the container will:

1. Accept SMTP connections on ports 25, 465, and 587
2. Check if the connecting IP is in the `allowed_hosts` list
3. If allowed, relay the email to Microsoft 365
4. If not allowed, reject the connection
5. Log all activity via rsyslog

### Testing

You can test the relay by sending an email from an allowed host:

```bash
telnet your-server-ip 25
HELO example.com
MAIL FROM: <sender@example.com>
RCPT TO: <recipient@example.com>
DATA
Subject: Test Email

This is a test message.
.
QUIT
```

## Security Considerations

- Only allow trusted networks in `allowed_hosts`
- Use TLS encryption for email transmission
- Regularly update the Docker image
- Monitor logs for unauthorized access attempts
- Consider firewall rules to restrict access to SMTP ports

## Troubleshooting

### Check container logs
```bash
docker-compose logs opensmtpd
```

### Check OpenSMTPD status
```bash
docker exec opensmtpd supervisorctl status
```

### Verify configuration
```bash
docker exec opensmtpd smtpd -n
```

## License

See LICENSE file for details.
