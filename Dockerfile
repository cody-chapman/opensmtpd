FROM debian:13-slim

# Set non-interactive mode
ENV DEBIAN_FRONTEND=noninteractive

# Install required packages
RUN apt-get update && apt-get install -y \
    opensmtpd \
    supervisor \
    && rm -rf /var/lib/apt/lists/*

# Create necessary directories
RUN mkdir -p /var/spool/opensmtpd \
    /var/log/opensmtpd

RUN touch /var/log/opensmtpd/opensmtpd.log
RUN touch /var/log/opensmtpd/opensmtpd.err.log

# Set proper permissions
RUN chmod 644 /etc/smtpd.conf

COPY entrypoint.sh /entrypoint.sh

# Expose SMTP ports
EXPOSE 25 587 465

# Start supervisor
CMD ["sh", "/entrypoint.sh"]
