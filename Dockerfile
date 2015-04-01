FROM google/golang
MAINTAINER Nicholas Hilem <nhilem@gmail.com>

WORKDIR /gopath/src/gadder
ADD . /gopath/src/gadder/
RUN go get gadder

EXPOSE 3000
# CMD []
CMD ["/gopath/bin/gadder"]
# ENTRYPOINT ["/gopath/bin/gadder"]
