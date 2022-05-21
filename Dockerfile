FROM ghcr.io/notional-labs/images/gno as build

RUN mkdir /opt/src /opt/build
ADD . /opt/src/
ADD . /opt/build/
WORKDIR /opt/build
RUN make


FROM ghcr.io/notional-labs/images/gno

COPY --from=build /opt/build/build/* /opt/gno/bin/
COPY --from=build /opt/src /opt/gno/src
ENV PATH="${PATH}:/opt/gno/bin"
WORKDIR /opt/gno/src
