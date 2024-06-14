FROM golang:1.22 as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
ENV GOPROXY=https://proxy.golang.org
RUN make clean && make cortextool

FROM       alpine:3.14
RUN        apk add --update --no-cache ca-certificates tzdata
COPY       --from=build /build_dir/cmd/cortextool/cortextool /usr/bin/cortextool
EXPOSE     80
ENTRYPOINT [ "/usr/bin/cortextool" ]
