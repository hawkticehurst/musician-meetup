FROM alpine
RUN apk add --no-cache ca-certificates
COPY summary /summary
EXPOSE 80
ENTRYPOINT ["/summary"]