#!/usr/bin/env bash
REPO_NAME="/terraform-provider-cf"
NAME="/terraform-provider-cloudfoundry"
OS=""
OWNER="cloudfoundry-community"
: "${TMPDIR:=${TMP:-$(
  CDPATH=/var:/
  cd -P tmp
)}}"
cd -- "${TMPDIR:?NO TEMP DIRECTORY FOUND!}" || exit
cd -

which terraform &>/dev/null
if [[ "$?" != "0" ]]; then
  echo "you must have terraform installed"
fi

if [[ "x$PROVIDER_CF_VERSION" == "x" ]]; then
  VERSION=$(curl -s https://api.github.com/repos/${OWNER}/${REPO_NAME}/releases/latest | grep tag_name | head -n 1 | cut -d '"' -f 4)
else
  VERSION=$PROVIDER_CF_VERSION
fi

echo "Installing ${NAME}-${VERSION}..."
if [[ "$OSTYPE" == "linux-gnu" || "$(uname -s)" == "Linux" ]]; then
  OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
  OS="darwin"
elif [[ "$OSTYPE" == "cygwin" ]]; then
  OS="windows"
elif [[ "$OSTYPE" == "msys" ]]; then
  OS="windows"
elif [[ "$OSTYPE" == "win32" ]]; then
  OS="windows"
else
  echo "Os not supported by install script"
  exit 1
fi

ARCHNUM=$(getconf LONG_BIT)
ARCH=""
CPUINFO=$(uname -m)
if [[ "$ARCHNUM" == "32" ]]; then
  ARCH="386"
else
  ARCH="amd64"
fi
if [[ "$CPUINFO" == "arm"* ]]; then
  ARCH="arm"
fi
FILENAME="${NAME}_${OS}_${ARCH}"
if [[ "$OS" == "windows" ]]; then
  FILENAME="${FILENAME}.exe"
fi

LINK="https://github.com/${OWNER}/${REPO_NAME}/releases/download/${VERSION}/${FILENAME}"
if [[ "$OS" == "windows" ]]; then
  FILEOUTPUT="${FILENAME}"
else
  FILEOUTPUT="${TMPDIR}/${FILENAME}"
fi
RESPONSE=200
if hash curl 2>/dev/null; then
  RESPONSE=$(curl --write-out %{http_code} -L -o "${FILEOUTPUT}" "$LINK")
else
  wget -o "${FILEOUTPUT}" "$LINK"
  RESPONSE=$?
fi

if [ "$RESPONSE" != "200" ] && [ "$RESPONSE" != "0" ]; then
  echo "File ${LINK} not found, so it can't be downloaded."
  rm "$FILEOUTPUT"
  exit 1
fi

chmod +x "$FILEOUTPUT"
mkdir -p ${HOME}/.terraform.d/plugins/${OS}_${ARCH}
if [[ "$OS" == "windows" ]]; then
  mv "$FILEOUTPUT" "${HOME}/.terraform.d/plugins/${OS}_${ARCH}/${NAME}"
else
  mv "$FILEOUTPUT" "${HOME}/.terraform.d/plugins/${OS}_${ARCH}/${NAME}"
fi

echo "${NAME}-${VERSION} has been installed."
