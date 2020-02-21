FROM golang:1.12.5-stretch as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
ENV GOPROXY=https://proxy.golang.org
RUN make clean && make chunktool

FROM       alpine:3.9
RUN        apk add --update --no-cache ca-certificates
COPY       --from=build /build_dir/cmd/chunktool/chunktool /usr/bin/chunktool
EXPOSE     80
ENTRYPOINT [ "/usr/bin/chunktool" ]
