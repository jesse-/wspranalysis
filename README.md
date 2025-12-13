# wspranalysis #
Utility for analysing WSPR data from wspr.live to assess transmitter &amp; antenna effectiveness.

This is my first Go program :)

## Build ##

To build the wspranalysis executable, just run `go build .` in the repo.

## Usage ##

To see usage information, run `./wspranalysis -h`. Required arguments are the target callsign (of the station whose effectiveness is to be assessed) and the band (e.g. 20m, 15m etc.).

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