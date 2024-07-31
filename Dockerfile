FROM alpine
RUN apk add libcap

COPY skipctl /skipctl
RUN setcap cap_net_raw+ep /skipctl

USER 150:150
ENTRYPOINT ["/skipctl"]
