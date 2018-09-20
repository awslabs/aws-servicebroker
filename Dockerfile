FROM deis/go-dev as builder

ENV PROJECT_DIR=/go/src/github.com/awslabs/aws-servicebroker
RUN mkdir -p $PROJECT_DIR
WORKDIR $PROJECT_DIR
ARG SOURCE_DIR="./"

COPY $SOURCE_DIR .

RUN dep ensure && make test && make linux

FROM alpine:latest

RUN apk add --no-cache ca-certificates bash

COPY --from=builder /go/src/github.com/awslabs/aws-servicebroker/servicebroker-linux /usr/local/bin/aws-servicebroker
COPY --from=builder /go/src/github.com/awslabs/aws-servicebroker/scripts/start_broker.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/start_broker.sh

CMD ["start_broker.sh"]
