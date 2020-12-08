FROM scratch

COPY ./bin/email-pipeline /go/bin/email-pipeline

ENTRYPOINT ["/go/bin/email-pipeline"]