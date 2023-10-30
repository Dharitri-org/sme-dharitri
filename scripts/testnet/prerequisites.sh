#!/usr/bin/env bash

export DHARITRITESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$DHARITRITESTNETSCRIPTSDIR/variables.sh"

export DISTRIBUTION=$(cat /etc/os-release | grep "^ID=" | sed 's/ID=//')


if [[ "$DISTRIBUTION" =~ ^(fedora|centos|rhel)$ ]]; then
  export PACKAGE_MANAGER="dnf"
  export REQUIRED_PACKAGES="git golang gcc lsof jq curl"
  echo "Using DNF to install required packages: $REQUIRED_PACKAGES"
fi

if [[ "$DISTRIBUTION" =~ ^(ubuntu|debian)$ ]]; then
  export PACKAGE_MANAGER="apt-get"
  export REQUIRED_PACKAGES="git gcc lsof jq curl"

  echo "Using APT to install required packages: $REQUIRED_PACKAGES"
fi

sudo $PACKAGE_MANAGER install -y $REQUIRED_PACKAGES

if [[ "$DISTRIBUTION" =~ ^(ubuntu|debian)$ ]]; then

  if ! [ -x "$(command -v go)" ]; then
    echo "Installing Go..."
    GO_LATEST=$(curl -sS https://golang.org/VERSION?m=text) 
    wget https://dl.google.com/go/$GO_LATEST.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf $GO_LATEST.linux-amd64.tar.gz
    rm $GO_LATEST.linux-amd64.tar.gz

    export GOROOT="/usr/local/go"
    export GOBIN="$HOME/go/bin"
    export PATH=$PATH:$GOROOT/bin:$GOBIN
    mkdir -p $GOBIN

    echo "export GOROOT=/usr/local/go" >> ~/.profile
    echo "export GOBIN=$HOME/go/bin" >> ~/.profile
    echo "export PATH=$PATH:$GOROOT/bin:$GOBIN" >> ~/.profile
    source ~/.profile 
  fi
fi


cd $(dirname $DHARITRIDIR)
git clone git@github.com:Dharitri-org/dharitri-deploy-go.git

if [ $PRIVATE_REPOS -eq 1 ]; then
  git clone git@github.com:Dharitri-org/dharitri-proxy-go.git

  git clone git@github.com:Dharitri-org/dharitri-txgen-go.git
  cd dharitri-txgen-go
  git checkout master
fi
