FROM       alpine:3.14
RUN        apk add --update --no-cache ca-certificates
COPY       benchtool /usr/bin/benchtool
EXPOSE     80
ENTRYPOINT [ "/usr/bin/benchtool" ]
