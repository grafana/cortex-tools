FROM       alpine:3.23@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11
RUN        apk add --update --no-cache ca-certificates
COPY       cortextool /usr/bin/cortextool
EXPOSE     80
ENTRYPOINT [ "/usr/bin/cortextool" ]
