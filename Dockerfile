FROM golang:alpine
RUN apk add --update \
    wget \
    git \
    && rm -rf /var/cache/apk/*
# RUN adduser -S -D -H -h /go/src/github.com/mrfleap/custom-compression appuser
# USER appuser
# RUN mkdir /go/src/github.com/mrfleap/custom-compression
ADD . /go/src/github.com/mrfleap/custom-compression
WORKDIR /go/src/github.com/mrfleap/custom-compression
RUN go get
RUN go install
RUN go build -o main .
RUN wget "https://data.cityofnewyork.us/api/views/zt9s-n5aj/rows.json?accessType=DOWNLOAD" -O rows.json
RUN wget "https://data.cityofchicago.org/api/geospatial/bbvz-uum9?method=export&format=Shapefile" -O boundaries.zip
CMD ["./main compress rows.json"]
# CMD ["./main compress boundaries.zip"]