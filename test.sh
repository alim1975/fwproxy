#!/bin/bash

# Build application, assuming GO is already installed
# It is also assumed that make is avaiable
export GOPATH=$(dirname $0)
make all

# Start the backend
echo "Starting URL DB backend..."
${GOPATH}/bin/dbase --host 127.0.0.1 --port 8888 --prefix /urlinfo/1/ --fwdb ${GOPATH}/src/db/main/filterurls.txt &
if [ $? -ne 0 ]; then
  echo "Failed to start URL DB backend..."
  exit 1
fi

echo "Done"

echo "Starting the proxy..."
${GOPATH}/bin/proxy --conf ${GOPATH}/src/proxy/main/proxy.yaml &
if [ $? -ne 0 ]; then
  echo "Failed to start the proxy..."
  killall -9 dbase
  exit 2
fi
echo "Done"
sleep 1

ret=$(curl -I -x http://127.0.0.1:8080 http://www.google.com 2>/dev/null | head -n 1 | cut -d$' ' -f2)
if [ "${ret}" != "200" ]; then
  echo "ERROR: expecting HTTP code: 200 got: ${ret}"
else
  ret=$(curl -I -x http://127.0.0.1:8080 http://www.pornhub.com 2>/dev/null | head -n 1 | cut -d$' ' -f2)
  if [ "${ret}" != "403" ]; then
    echo "ERROR: expecting HTTP code: 403 got: ${ret}"
  fi
fi

# Kill URL DB service
killall -9 dbase

# Kill the proxy
killall -9 proxy

