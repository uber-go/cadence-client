FROM golang:1.19.5

RUN mkdir -p /go/src/go.uber.org/cadence
WORKDIR /go/src/go.uber.org/cadence

ADD go.mod go.sum /go/src/go.uber.org/cadence/

# allow git-status and similar to work
RUN git config --global --add safe.directory /go/src/go.uber.org/cadence
RUN GO111MODULE=on go mod download

# Install dependencies needed to install buildkite-agent
RUN apt-get install -y apt-transport-https dirmngr curl

# Add Buildkite's GPG key
RUN curl -fsSL https://keys.openpgp.org/vks/v1/by-fingerprint/32A37959C2FA5C3C99EFBC32A79206696452D198 | \
  gpg --dearmor -o /usr/share/keyrings/buildkite-agent-archive-keyring.gpg

# Add the repository to Apt sources
RUN echo \
  "deb [signed-by=/usr/share/keyrings/buildkite-agent-archive-keyring.gpg] https://apt.buildkite.com/buildkite-agent stable main" | \
  tee /etc/apt/sources.list.d/buildkite-agent.list

# Install Buildkite agent
RUN apt-get update && apt-get install -yy --no-install-recommends buildkite-agent
