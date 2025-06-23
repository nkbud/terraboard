FROM golang:1.21 as builder
WORKDIR /opt/build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN make build

FROM node:lts as node-builder
WORKDIR /opt/build/terraboard-vuejs
COPY static/terraboard-vuejs/package.json static/terraboard-vuejs/yarn.lock ./
RUN yarn install
COPY static/terraboard-vuejs/ ./
RUN yarn run build

FROM scratch
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /opt/build/terraboard /
COPY --from=node-builder /opt/build/terraboard-vuejs/dist /static
ENTRYPOINT ["/terraboard"]
CMD [""]
