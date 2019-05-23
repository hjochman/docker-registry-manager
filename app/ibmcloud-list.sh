#!/bin/sh
# $1=password $2=host

set -e
function join_by { local IFS="$1"; shift; echo "$*"; }

BLUEMIX_API_KEY=$1

ibmcloud config --check-version=false

# login to IBM Cloud via credentials provided via (encrypted) environment
# variables
bluemix login \
  --apikey "${BLUEMIX_API_KEY}" \
  -a "https://api.eu-de.bluemix.net" &> /dev/null
#  -o "${BLUEMIX_ORGANIZATION}" \
#  -s "${BLUEMIX_SPACE}"

repos=$(ibmcloud cr image-list --format "{{ if gt .Size 1000000 }}{{ .Repository }}{{end}}" | cut -d "/" -f2-100| uniq)
join_by , ${repos}

ibmcloud logout &> /dev/null
