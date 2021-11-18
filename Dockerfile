FROM ubuntu:latest
LABEL maintainer="matt@matthewrogers.org"

ENV HOME /root
ENV LC_ALL C.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US.UTF-8
ENV GOPATH /root/go
ENV TZ America/New_York
ENV DEBIAN_FRONTEND noninteractive

ADD activerecon.linux /bin/activerecon
ADD settings.scan /root/settings.scan
ADD static /root/static
ADD visual /root/visual
ADD js /root/js
ADD css /root/css
ADD run.sh /root/run.sh


RUN chmod +x /root/run.sh
RUN chmod +x /bin/activerecon
RUN apt update
RUN apt install wget -y
RUN apt install masscan -y
RUN apt install libpcap-dev -y
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN apt install ./google-chrome-stable_current_amd64.deb -y
RUN wget https://github.com/sensepost/gowitness/releases/download/2.3.6/gowitness-2.3.6-linux-amd64 -O /bin/gowitness
RUN chmod +x /bin/gowitness

EXPOSE 9009/tcp

# Run this thing
CMD ["/root/run.sh"]

#Example Invocation
#