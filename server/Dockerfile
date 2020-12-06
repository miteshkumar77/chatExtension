FROM golang:1.15.5

WORKDIR /go/src/app
COPY . .

RUN make build

ENV THREADS=3
ENV RATELIMIT=2
ENV PORT=8080

EXPOSE ${PORT}

CMD make run_build
