#!/bin/sh

IMG_VERSION=${IMG_VERSION?"Define IMG_VERSION env"}

yq -i ".images[] |= sub(\":.*\", \":${IMG_VERSION}\")" component-config.yaml
