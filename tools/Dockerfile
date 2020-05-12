FROM golang:1.14

RUN apt-get update && apt-get install -y curl wget gnupg2
RUN wget https://dl.yarnpkg.com/debian/pubkey.gpg && \
    apt-key add pubkey.gpg && \
    rm pubkey.gpg && \
    echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN curl -sL https://deb.nodesource.com/setup_10.x |  bash -
RUN apt-get update
RUN apt-get install -y gettext-base git python3 python3-pip python3-yaml shellcheck yarn build-essential openssl libssl-dev nodejs
RUN pip3 install pytablewriter==0.51.0
RUN npm install --global prettier@2.0.5

RUN go get golang.org/x/tools/cmd/goimports
