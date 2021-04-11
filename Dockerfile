FROM golang:1.16

# set a directory for the app
WORKDIR /go/src/app
#copy all the files to the container
COPY . .

RUN go get -d -v ./...
RUN go install -v ./..

CMD ["app"]