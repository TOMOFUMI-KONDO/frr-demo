#!/bin/sh

route add default gw 172.36.0.2
route del default gw 172.36.0.1

tail -f /dev/null
