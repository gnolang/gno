FROM ghcr.io/faddat/cosmos as build

RUN mkdir /opt/src /opt/build
ADD . /opt/src/
ADD . /opt/build/
WORKDIR /opt/build
RUN make

# Hard to say if we want one of these minimal runtime containers: it defies docker idiom but I've seen runtime containers cause serious trouble on production chains. 
FROM ghcr.io/faddat/archlinux

COPY --from=build /opt/build/build/* /opt/gno/bin/
COPY --from=build /opt/src /opt/gno/src
ENV PATH="${PATH}:/opt/gno/bin"
WORKDIR /opt/gno/src
