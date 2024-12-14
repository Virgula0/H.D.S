FROM golang:1.22.2

# update
RUN apt update && \
    apt upgrade -y && \
    apt install make libxml2-dev libxmlsec1-dev pkg-config -y

# COPY THE CONTENT OF BACKEND INSIDE APP
WORKDIR /app/backend

RUN adduser backend && \
    chown -R backend:backend .

USER backend
COPY --chown=backend . .

# go modules
# COPY go.mod go.sum ./
RUN go mod download && \
    # go mod verify && \ don't know if this updates packages, we don't want it if yes
    go mod tidy &&\
    make build

CMD build/server