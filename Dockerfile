# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.17.8 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-mwe
COPY . .
RUN make install

############# gardener-extension-mwe
FROM alpine:3.15.0 AS gardener-extension-mwe

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-mwe /gardener-extension-mwe
ENTRYPOINT ["/gardener-extension-mwe"]
