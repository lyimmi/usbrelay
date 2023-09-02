#!/bin/bash

PCAPPATH="/sys/devices/pci0000:00/0000:00:01.2/0000:02:00.0/0000:03:08.0/0000:06:00.1/usb1/1-3"

commands=(
"list"
"on REL01 all"
"off REL01 all"
"toggle REL01 all"
"setserial REL01 REL02"
"setserial REL02 REL01"
"list"
)
tests=(
"umockdev-run --device relay.umockdev --pcap $PCAPPATH=01-list-start.pcapng /tmp/usbrelay list"
"umockdev-run --device relay.umockdev --pcap $PCAPPATH=02-on-all.pcapng /tmp/usbrelay on REL01 all"
"umockdev-run --device relay.umockdev --pcap $PCAPPATH=03-off-all.pcapng /tmp/usbrelay off REL01 all"
"umockdev-run --device relay.umockdev --pcap $PCAPPATH=04-toggle-all.pcapng /tmp/usbrelay toggle REL01 all"
"umockdev-run --device relay.umockdev --pcap $PCAPPATH=05-setserial-rel01-rel02.pcapng /tmp/usbrelay setserial REL01 REL02"
"umockdev-run --device relay.umockdev --pcap $PCAPPATH=06-setserial-rel02-rel01.pcapng /tmp/usbrelay setserial REL02 REL01"
"umockdev-run --device relay.umockdev --pcap $PCAPPATH=07-list-end.pcapng /tmp/usbrelay list"
)

for (( i = 0; i < ${#tests[@]} ; i++ )); do
    printf 'Running: %s\n' "${commands[$i]}"

    if ! ${tests[$i]}; then
      printf '\Failed: %s\n' "${commands[$i]}"
      exit 1
    fi
done