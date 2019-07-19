FROM golang:1.12.5-stretch as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
RUN ls 
RUN make clean && make cortex-tool

FROM       alpine:3.9
RUN        apk add --update --no-cache ca-certificates
COPY       --from=build /cortex-tool/cmd/cortex-tool/cortex-tool /usr/bin/cortex-tool
EXPOSE     80
ENTRYPOINT [ "/usr/bin/cortex-tool" ]
