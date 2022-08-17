#!/bin/bash

set -ex

if [[ "3" == `ls -la ./node_modules/ | wc -l` ]]; then npm install --production=false; fi && \
npm run build:prod --registry=https://registry.npm.taobao.org