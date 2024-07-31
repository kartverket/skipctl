FROM scratch
COPY skipctl /skipctl
ENTRYPOINT ["/skipctl"]
