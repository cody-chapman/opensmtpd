FROM debian:12-slim

# Set non-interactive mode
ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl ca-certificates gnupg vim supervisor opensmtpd golang \
    && rm -rf /var/lib/apt/lists/*
    
WORKDIR /app

COPY smtpd-ui/ ./smtpd-ui/
WORKDIR /app/smtpd-ui 
RUN go build -o smtpd-ui .  

# Create necessary directories
RUN mkdir -p /var/spool/opensmtpd \
    /var/log/opensmtpd

WORKDIR /app

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
RUN chmod 644 /etc/smtpd.conf

# Expose SMTP ports
EXPOSE 25 587 465 8080 8085

# Start the entrypoint
ENTRYPOINT ["/entrypoint.sh"]
