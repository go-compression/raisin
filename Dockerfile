FROM golang:alpine
RUN apk add --update \
    wget \
    unzip \
    git \
    && rm -rf /var/cache/apk/*
# RUN adduser -S -D -H -h /go/src/github.com/go-compression/raisin appuser
# USER appuser
# RUN mkdir /go/src/github.com/go-compression/raisin
ADD . /go/src/github.com/go-compression/raisin
WORKDIR /go/src/github.com/go-compression/raisin
RUN go get
RUN go install
RUN go build -o main .
# RUN wget "https://data.wprdc.org/dataset/9e0ce87d-07b8-420c-a8aa-9de6104f61d6/resource/96474373-bcdb-42cf-af5d-3683e326e227/download/sales-validation-codes-dictionary.pdf" -O sales.pdf
# RUN wget "https://data.cityofnewyork.us/api/views/zt9s-n5aj/rows.json?accessType=DOWNLOAD" -O rows.json
# RUN wget "https://data.cityofchicago.org/api/geospatial/bbvz-uum9?method=export&format=Shapefile" -O boundaries.zip
RUN wget "http://corpus.canterbury.ac.nz/resources/cantrbry.zip" -O canterbury.zip
RUN unzip canterbury.zip -d ./