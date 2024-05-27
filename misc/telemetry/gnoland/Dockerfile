# Use the existing gno image as the base image
FROM ghcr.io/gnolang/gno/gnoland:latest AS base

# Copy the setup script into the container
COPY ./setup.sh .

# Make the script executable
RUN chmod +x ./setup.sh

# Run the setup
ENTRYPOINT ["sh"]

CMD ["./setup.sh"]