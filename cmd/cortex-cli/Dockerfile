FROM golang:1.12.5-stretch as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
RUN ls 
RUN make clean && make cortex-cli

FROM       alpine:3.9
RUN        apk add --update --no-cache ca-certificates
COPY       --from=build /cortex-cli/cmd/cortex-cli/cortex-cli /usr/bin/cortex-cli
EXPOSE     80
ENTRYPOINT [ "/usr/bin/cortex-cli" ]
