# wspranalysis #
Utility for analysing WSPR data from wspr.live to assess transmitter &amp; antenna effectiveness.

This is my first Go program :)

## Build & Install ##

Build the CLI binary from the `cmd` directory:

```bash
go build ./cmd/wspranalysis
```

Or install it into your Go `bin` directory:

```bash
go install github.com/jesse-/wspranalysis/cmd/wspranalysis@latest
```

Run tests:

```bash
go test ./...
```

You can also run the tool directly with `go run` while developing, eg:

```bash
go run ./cmd/wspranalysis -h
```

## Usage ##

After building, run the CLI binary. The required arguments are the target callsign (the station to assess) and the band (for example `20m`, `15m`, etc.). Example:

```bash
./wspranalysis K1ABC 20m
```

Common flags:

- `-start` : RFC3339 start time for the query (default: 24h ago)
- `-duration` : duration to analyse (e.g. `24h`, `30m`)
- `-norm` : transmit power in dBm to normalise SNRs (default: 43)
- `-v` : verbose output (lists all transmitters heard by each receiver)

## What the Tool Does ##

[WSPR](https://www.arrl.org/wspr) is an amateur radio mode used for testing signal propagation. This tool makes use of the [wspr.live](https://wspr.live) database of WSPR reception reports to analyse the performance of a selected target transmitter.

The transmitting station's perfomance (which, in most cases, will be a function of the antenna and its location) is assessed by ranking its signal against that of other transmitters at each receiver. To ensure a valid comparison, the following steps are taken:

* At each receiver, the signal from the target transmitter is only compared against other transmitters which are a similar distance from the receiver.
* All the SNRs are normalised with respect to the reported transmit power. I.e. all the SNRs are as they would be if every transmitter used the same nominal power.

The offset of the SNR of the target transmitter from the median of all comparable signals received is shown for each receiving station. Another metric is generated across all receivers as follows:

1. For each receiving station, all the normalised SNRs (except that of the target signal itself) are recorded relative to the signal from the target transmitter.
2. These relative SNRs are aggregated across all receivers.
3. The final metric is the negated median (in dB) of these aggregated relative SNRs.

In other words, the final metric is a dB figure which reflects how the signal from the target transmitter generally compares with the signal from other transmitters (once range and power differences have been taken into account). A positive value means that the target generally outperforms others while a negative value means that it underperforms. One should note the following though:

* A generally underperforming antenna might actually perform very well in specific directions over specific propagation paths.
* Because of power normalisation, a high performing very low power station may only perform well in real life if it uses a more typical transmit power.
