FROM debian:sid
ENV LANG C.UTF-8
ENV DEBIAN_FRONTEND=noninteractive

RUN echo 'root:devnet-gui-test' | chpasswd
RUN apt-get update && apt-get install -y \
    golang \
    libgtk-3-dev \
    libcairo2-dev \
    libglib2.0-dev \
    gstreamer1.0-x \
    gstreamer1.0-gtk3 \
    gstreamer1.0-libav \
    gstreamer1.0-tools \
    gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-ugly \
    gstreamer1.0-plugins-bad \
    libgstreamer1.0-dev \
    libgstreamer-plugins-base1.0-dev \
    xrdp \
    openbox \
    xterm \
    && rm -rf /var/lib/apt/lists/*

ENV GOPATH=/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV GO111MODULE=on

WORKDIR /go/src/devnet

RUN go get github.com/mjibson/esc; \
    go get github.com/gotk3/gotk3/gtk 

COPY go.mod go.sum ./
RUN go mod download

COPY scripts/docker/xsession /root/.xsession

COPY . .

WORKDIR /go/src/devnet/cmd/devnet
RUN GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/devnet

COPY configs/docker/client.yaml /root/.config/devnet/config.yaml

CMD ["sh", "-c", "/etc/init.d/xrdp start && tail -f /dev/null"]
