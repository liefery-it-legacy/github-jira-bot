FROM golang:1.12.4 AS builder

ENV GIT_SERVER=github.com
ENV GIT_ORG=Benbentwo
ENV GIT_REPO=github-jira-bot
# Build arguments
ARG binary_name=github-jira-bot
    # See ./sample-data/go-os-arch.csv for a table of OS & Architecture for your base image
ARG target_os=linux
ARG target_arch=amd64

# Build the server Binary
WORKDIR /go/src/${GIT_SERVER}/${GIT_ORG}/${GIT_REPO}
ADD . ./
RUN go get -u ./...
RUN CGO_ENABLED=0 GOOS=${target_os} GOARCH=${target_arch} go build -a -o /app/${binary_name} main.go

RUN ls /app

#-----------------------------------------------------------------------------------------------------------------------

FROM centos:7

LABEL author="Benjamin Smith"
COPY --from=builder ./app/github-jira-bot /usr/bin/github-jira-bot
RUN ["chmod", "-R", "+x", "/usr/bin/github-jira-bot"]

#ENTRYPOINT ["github-jira-bot", "--help"]
ENTRYPOINT ["tail", "-f", "/dev/null"]
