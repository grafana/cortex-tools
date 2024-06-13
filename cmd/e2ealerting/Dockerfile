FROM golang:1.22 as build
ARG GOARCH="amd64"
COPY . /build_dir
WORKDIR /build_dir
RUN make clean && make e2ealerting

FROM       alpine:3.14
RUN        apk add --update --no-cache ca-certificates
COPY       --from=build /build_dir/cmd/e2ealerting/e2ealerting /usr/bin/e2ealerting
EXPOSE     80
ENTRYPOINT [ "/usr/bin/e2ealerting" ]
