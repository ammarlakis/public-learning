
FROM ubuntu:latest
RUN apt-get update && apt-get install -y     vim
CMD ["docker", "version", "--format", "'{{json .Client.Version}}'"]
