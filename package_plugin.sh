#!/bin/bash

set -ex

export GRAFANA_PLUGIN_ID=$(cat dist/plugin.json | jq -r .id)
export GRAFANA_PLUGIN_VERSION=$(cat dist/plugin.json | jq -r .info.version)
export GRAFANA_PLUGIN_TYPE=$(cat dist/plugin.json | jq -r .type)
export GRAFANA_PLUGIN_ARTIFACT=${GRAFANA_PLUGIN_ID}-${GRAFANA_PLUGIN_VERSION}.zip
export GRAFANA_PLUGIN_ARTIFACT_CHECKSUM=${GRAFANA_PLUGIN_ARTIFACT}.md5

mkdir -p /tmp/${GRAFANA_PLUGIN_ID}
cp -rfv dist/* /tmp/${GRAFANA_PLUGIN_ID}/
pushd /tmp/
zip ${GRAFANA_PLUGIN_ARTIFACT} ${GRAFANA_PLUGIN_ID} -r
md5sum ${GRAFANA_PLUGIN_ARTIFACT} > /tmp/${GRAFANA_PLUGIN_ARTIFACT_CHECKSUM}
popd
