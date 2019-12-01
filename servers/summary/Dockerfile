FROM alpine
RUN apk add --no-cache ca-certificates
COPY summary /summary
EXPOSE 443
ENTRYPOINT ["/summary"]