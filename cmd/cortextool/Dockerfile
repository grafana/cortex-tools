FROM golang:1.22@sha256:1cf6c45ba39db9fd6db16922041d074a63c935556a05c5ccb62d181034df7f02 as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
ENV GOPROXY=https://proxy.golang.org
RUN make clean && make cortextool

FROM       alpine:3.14@sha256:0f2d5c38dd7a4f4f733e688e3a6733cb5ab1ac6e3cb4603a5dd564e5bfb80eed
RUN        apk add --update --no-cache ca-certificates tzdata
COPY       --from=build /build_dir/cmd/cortextool/cortextool /usr/bin/cortextool
EXPOSE     80
ENTRYPOINT [ "/usr/bin/cortextool" ]
