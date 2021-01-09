
#
# Alice - The friendly BGP looking glass
#

# Build frontend first
FROM node:11 AS frontend

# Install dependencies 
WORKDIR /src/alice-lg/client
ADD client/package.json .
ADD client/yarn.lock .

RUN npm install -g gulp@4.0.0
RUN npm install -g gulp-cli
RUN yarn install

# Add frontend
WORKDIR /src/alice-lg/client
ADD client .

# Build frontend
RUN DISABLE_LOGGING=1 NODE_ENV=production /usr/local/bin/gulp

# Build the backend
FROM golang:1.15 AS backend

# Install dependencies
WORKDIR /src/alice-lg
ADD go.mod .
ADD go.sum .
RUN go mod download
RUN go install github.com/GeertJohan/go.rice/rice

# Add client
COPY --from=frontend /src/alice-lg/client/build client/build

# Build backend
WORKDIR /src/alice-lg
ADD VERSION .
COPY etc etc

WORKDIR /src/alice-lg/backend
ADD backend .
RUN rice embed-go

# RUN go build -o alice-lg-linux-amd64 -ldflags="-X main.version=4.0.3"
RUN make alpine

FROM alpine:latest
COPY --from=backend /src/alice-lg/backend/alice-lg-linux-amd64 /usr/bin/alice-lg
COPY --from=backend /src/alice-lg/etc/alice-lg/alice.docker.conf /etc/alice-lg/alice.conf
RUN ls -lsha /usr/bin/alice-lg
#ENV GODEBUG=madvdontneed=1,gctrace=1,scavenge=1,scavtrace=1
ENV GODEBUG=madvdontneed=1,gctrace=1

EXPOSE 6060:6060
CMD ["/usr/bin/alice-lg"]
