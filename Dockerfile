FROM debian:13-slim

# Set non-interactive mode
ENV DEBIAN_FRONTEND=noninteractive

# Install required packages
RUN apt-get update && apt-get install -y \
    opensmtpd \
    supervisor \
    whiptail \
    && rm -rf /var/lib/apt/lists/*

# Create necessary directories
RUN mkdir -p /var/spool/opensmtpd \
    /var/log/opensmtpd


# Set proper permissions
RUN chmod 644 /etc/smtpd.conf

COPY smtpd-manage /usr/bin/smtpd-manage
RUN chmod +x /usr/bin/smtpd-manage
COPY entrypoint.sh /entrypoint.sh

# Expose SMTP ports
EXPOSE 25 587 465

# Start supervisor
CMD ["sh", "/entrypoint.sh"]
