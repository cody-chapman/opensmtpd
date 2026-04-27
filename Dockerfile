FROM debian:12-slim

# Set non-interactive mode
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl ca-certificates gnupg vim supervisor whiptail opensmtpd \
    && rm -rf /var/lib/apt/lists/*
    
# Add Microsoft repository and install PowerShell
RUN curl https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > /etc/apt/keyrings/microsoft.gpg \
    && echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/microsoft.gpg] https://packages.microsoft.com/debian/12/prod bookworm main" > /etc/apt/sources.list.d/microsoft.list \
    && apt-get update && apt-get install -y powershell \
    && rm -rf /var/lib/apt/lists/*

RUN /usr/bin/pwsh -c "Set-PSRepository -Name 'PSGallery' -InstallationPolicy Trusted; Install-Module -Name 'Pode' -Scope AllUsers -Force"

# Create necessary directories
RUN mkdir -p /var/spool/opensmtpd \
    /var/log/opensmtpd

WORKDIR /app

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
RUN chmod 644 /etc/smtpd.conf
RUN chmod +x /usr/bin/smtpd-manage
COPY entrypoint.sh /entrypoint.sh

# Expose SMTP ports
EXPOSE 25 587 465 8080 8085

# Start the entrypoint
ENTRYPOINT ["/entrypoint.sh"]
