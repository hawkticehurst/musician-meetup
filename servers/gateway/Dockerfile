FROM alpine
RUN apk add --no-cache ca-certificates
COPY gateway /gateway
EXPOSE 443
ENTRYPOINT ["/gateway"]
