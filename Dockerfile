FROM ubuntu:24.04

RUN apt update
RUN apt install -y curl wget xz-utils unzip

# RUN apt install -y gpac

# RUN apt install -y python3 pip
# RUN pip install -U openai-whisper

RUN curl -sL https://deb.nodesource.com/setup_18.x | bash -
RUN apt install -y nodejs
RUN npm install --global yarn

ENV FFMPEG_VERSION="ffmpeg-n6.0-latest-linux64-gpl-6.0"
RUN wget -q https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/${FFMPEG_VERSION}.tar.xz
RUN tar -xvf ${FFMPEG_VERSION}.tar.xz
RUN mv ${FFMPEG_VERSION}/bin/* /usr/bin/
RUN rm -rf ${FFMPEG_VERSION}*

WORKDIR /usr/local/app
COPY . .

RUN yarn install --frozen-lockfile --development
RUN yarn build

EXPOSE 5000

CMD yarn start
