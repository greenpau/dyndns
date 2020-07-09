#!/bin/bash
set -e
#set -x

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )";
PKG_NAME=dyndns
PKG_VERSION=$(cat ${ROOT_DIR}/../../VERSION | head -1)
PKG_RELEASE=$(cat ${ROOT_DIR}/RELEASE | head -1)
PKG_DISTRO=el7
PKG_CPU_ISA=x86_64
PKG_CPU_ARCH=amd64
PKG_OS=linux
PKG_RPM_FILE=${PKG_NAME}-${PKG_VERSION}-${PKG_RELEASE}.${PKG_DISTRO}.${PKG_CPU_ISA}
PKG_RPM_SPEC_FILE=${PKG_NAME}-${PKG_VERSION}-${PKG_RELEASE}.${PKG_DISTRO}.${PKG_CPU_ISA}.spec

cd ${ROOT_DIR}
gorpm --version
gorpm test --file config.json
if [ $? -eq 0 ]; then
    echo "INFO: Successfully validated configuration file"
else
    echo "ERROR: Failed to validate configuration file" >&2
fi

mkdir -p ./usr/local/bin
rm -rf ./usr/local/bin/${PKG_NAME}
cp ../../bin/${PKG_NAME} ./usr/local/bin/${PKG_NAME}
./usr/local/bin/${PKG_NAME} --version

rm -rf build
mkdir -p src

echo "INFO: Creating directories and override macros for rpmbuild"
rm -rf ~/rpmbuild/*
mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
echo '%_topdir %(echo $HOME)/rpmbuild' > ~/.rpmmacros

# Architectures: https://github.com/golang/go/blob/master/src/go/build/syslist.go
# Common architectures are: amd64, 386
#
# gorpm generate-spec --version 1.0.0 --file config.json --arch 386
# gorpm generate-spec --version 1.0.0 --file config.json --arch amd64

gorpm generate-spec \
  --file config.json \
  --arch "${PKG_CPU_ARCH}" \
  --version ${PKG_VERSION} \
  --release ${PKG_RELEASE} \
  --distro ${PKG_DISTRO} \
  --cpu ${PKG_CPU_ISA} \
  --output ./spec/${PKG_RPM_SPEC_FILE}

cd ${ROOT_DIR} && tar --strip-components 1 --owner=0 --group=0 \
  -czvf ~/rpmbuild/SOURCES/${PKG_RPM_FILE}.tar.gz \
  ./etc/sysconfig/${PKG_NAME}.conf \
  ./etc/profile.d/${PKG_NAME}.sh \
  ./lib/systemd/system/${PKG_NAME}.service \
  ./usr/lib/tmpfiles.d/${PKG_NAME}.conf \
  ./usr/local/bin/${PKG_NAME} \
  ./etc/${PKG_NAME}/config_template.json

tar -tvzf ~/rpmbuild/SOURCES/${PKG_RPM_FILE}.tar.gz

rpmbuild --nodeps --target ${PKG_CPU_ISA} -ba ./spec/${PKG_RPM_SPEC_FILE}
echo "INFO: list files in ~/rpmbuild/RPMS/${PKG_CPU_ISA}/${PKG_RPM_FILE}.rpm"
rpm -qlp ~/rpmbuild/RPMS/${PKG_CPU_ISA}/${PKG_RPM_FILE}.rpm

cd ${ROOT_DIR} && mkdir -p dist
rm -rf ./dist/${PKG_RPM_FILE}.rpm
cp ~/rpmbuild/RPMS/${PKG_CPU_ISA}/${PKG_RPM_FILE}.rpm ./dist/${PKG_RPM_FILE}.rpm
echo "SCP:       scp ./assets/rpm/dist/${PKG_RPM_FILE}.rpm root@remote:/tmp/"
echo "Install:   sudo yum -y localinstall ./assets/rpm/dist/${PKG_RPM_FILE}.rpm"
echo "RPM File:  ./assets/rpm/dist/${PKG_RPM_FILE}.rpm"

exit 0
