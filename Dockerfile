FROM debian:13-slim

# Set non-interactive mode
ENV DEBIAN_FRONTEND=noninteractive

# Install required packages
RUN apt-get update && apt-get install -y \
    opensmtpd \
    rsyslog \
    supervisor \
    && rm -rf /var/lib/apt/lists/*

# Create necessary directories
RUN mkdir -p /var/spool/opensmtpd \
    /var/log/supervisor \
    /var/log/rsyslog \
    /var/log/opensmtpd

# Copy configuration files
#COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
#COPY smtpd.conf /etc/smtpd.conf
#COPY rsyslog.conf /etc/rsyslog.conf
#COPY smtpd-rsyslog.conf /etc/rsyslog.d/10-smtpd.conf

# Set proper permissions
RUN chmod 644 /etc/smtpd.conf \
    && chmod 644 /etc/rsyslog.conf

COPY entrypoint.sh /entrypoint.sh

# Expose SMTP ports
EXPOSE 25 587 465

# Start supervisor
CMD ["sh", "/entrypoint.sh"]
