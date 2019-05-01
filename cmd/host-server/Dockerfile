from ubuntu

RUN mkdir -p /opt/thorium/host-server
ADD host-server /opt/thorium/host-server/host-server
ADD host.config /opt/thorium/host-server/host.config

WORKDIR /opt/thorium/host-server

CMD [ "./host-server" ]

EXPOSE 10000-60000
