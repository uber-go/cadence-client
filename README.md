# Go framework for Cadence [![Build Status](https://travis-ci.org/uber-go/cadence-client.svg?branch=master)](https://travis-ci.org/uber-go/cadence-client) [![Coverage Status](https://coveralls.io/repos/uber-go/cadence-client/badge.svg?branch=master&service=github)](https://coveralls.io/github/uber-go/cadence-client?branch=master)
[Cadence](https://github.com/uber/cadence) is a distributed, scalable, durable, and highly available orchestration engine we developed at Uber Engineering to execute asynchronous long-running business logic in a scalable and resilient way.

`cadence-client` is the framework for authoring workflows and activities in Go.

## Samples

For samples, see [Cadence Samples](https://github.com/samarabbas/cadence-samples).

## Run Cadence Server

Run Cadence Server using Docker Compose:

    curl -O https://raw.githubusercontent.com/uber/cadence/master/docker/docker-compose.yml
    docker-compose up

If this does not work, see instructions for running the Cadence Server at https://github.com/uber/cadence/blob/master/README.md.

## Contributing
We'd love your help in making Cadence-client great. Please review our [instructions](CONTRIBUTING.md).

## License
MIT License, please see [LICENSE](LICENSE) for details.
