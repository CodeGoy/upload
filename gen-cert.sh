#!/bin/bash

openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes -keyout server.key -out server.crt
