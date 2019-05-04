from ubuntu

EXPOSE 10000-60000

RUN apt-get update && apt-get install -y iputils-ping netcat dnsutils

RUN mkdir -p /opt/thorium/host-server
RUN mkdir /opt/thorium/host-server/config

WORKDIR /opt/thorium/host-server

ADD host-server /opt/thorium/host-server/host-server

CMD [ "./host-server" ]

