# OpenSMTPD

A modern Go-based implementation of an OpenSMTPD Docker relay that acts as a secure email relay to Microsoft 365, with configurable access control for allowed hosts.

## Overview

**OpenSMTPD** is a containerized SMTP relay server using OpenSMTPD that forwards emails to Microsoft 365's SMTP servers. It includes access control to restrict relaying to only specified IP addresses or networks, comprehensive logging, and a web-based UI for management.

## Features

- 📧 **SMTP Relay** - Secure email forwarding to Microsoft 365
- 🔒 **Access Control** - Restrict relaying to allowed hosts/networks
- 🐳 **Docker Containerized** - Easy deployment and portability
- 📝 **Comprehensive Logging** - Integrated rsyslog and supervisord logging
- 🌐 **Web UI** - Go-based management interface with `smtpd-ui`
- 🔧 **Configurable** - Easy-to-modify configuration files
- 🌐 **Multi-Port Support** - Standard SMTP (25), SMTPS (465), and Submission (587)

## Technology Stack

- **Go** - 97.9% (Primary language for UI and utilities)
- **Docker** - 1.3% (Container configuration)
- **Shell** - 0.8% (Setup and helper scripts)

**Language:** Go 1.23

## Prerequisites

- Docker
- Docker Compose
- Access to Microsoft 365 SMTP servers (typically `yourdomain.mail.protection.outlook.com`)
- Go 1.23+ (for local development)

## Quick Start

### Using Docker Compose

1. Clone this repository:
```bash
git clone https://github.com/cody-chapman/opensmtpd.git
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

### Local Development

```bash
cd smtpd-ui
go build -o smtpd-ui
./smtpd-ui
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TZ` | Timezone for logging | America/Chicago |
| `SMTPD_ADMIN_PORT` | Web UI port | 8080 |

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
6. Provide a web UI for management (accessible on port 8080 by default)

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
- Regularly update the Docker image and dependencies
- Monitor logs for unauthorized access attempts
- Consider firewall rules to restrict access to SMTP ports
- Secure the web UI with proper authentication and network isolation

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

### View web UI
```bash
# If running locally: http://localhost:8080
# If running on server: http://your-server-ip:8080
```

## Project Structure

```
.
├── smtpd-ui/              # Go-based web UI
│   └── go.mod             # Go module definition
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile             # Container image definition
├── smtpd.conf            # OpenSMTPD configuration
├── rsyslog.conf          # Logging configuration
├── supervisord.conf      # Process supervisor configuration
├── allowed_hosts         # List of allowed relay hosts
└── README.md             # This file
```

## License

This project is licensed under the GNU General Public License v2.0 - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues, questions, or suggestions, please open an issue on GitHub.
