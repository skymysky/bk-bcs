#!/bin/bash

module="bcs-gateway-discovery"

cd /data/bcs/${module}
chmod +x ${module}

#check configuration render
if [ $BCS_CONFIG_TYPE == "render" ]; then
  cd /data/bcs/${module}
  cat ${module}.json.template | envsubst | tee ${module}.json

  #apisix configuration
  cd /usr/local/apisix/conf
  cat config.yaml.template | envsubst | tee config.yaml
fi

apisix start

#setting tls certification by json
if [ "x${bcsSSLJSON}" == "x" ]; then
  echo "lost apisix bkbcs SSL json"
  exit 1
fi

curl http://127.0.0.1:8000/apisix/admin/ssl/bkbcs \
  -H"X-API-KEY: ${adminToken}" -X PUT -d@${bcsSSLJSON}

#starting module
cd /data/bcs/${module}
/data/bcs/${module}/${module} $@
