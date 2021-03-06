FROM golang:1.16.0-stretch as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
ENV GOPROXY=https://proxy.golang.org
RUN make clean && make benchtool

FROM       alpine:3.9
RUN        apk add --update --no-cache ca-certificates
COPY       --from=build /build_dir/cmd/benchtool/benchtool /usr/bin/benchtool
EXPOSE     80
ENTRYPOINT [ "/usr/bin/benchtool" ]
