#
# Harmony API DockerFile
#

FROM phusion/baseimage:0.9.16



#
# Misc Docker Config
#

# set the maintainer
MAINTAINER pmccarren

# Set correct environment variables.
ENV HOME /root

# Use baseimage-docker's init system.
CMD ["/sbin/my_init"]

# enter the container at home
WORKDIR /root



#
# Install packages
#

RUN apt-get update && \
	apt-get install -y wget && \
	wget -qO- https://get.docker.com/ | sh && \
	wget -q https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz && \
	tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz && \
	rm go1.4.2.linux-amd64.tar.gz


#
# Go
#
ENV PATH $PATH:/usr/local/go/bin
ENV GOPATH /gocode



#
# Misc Init
#
ADD docker_fs/etc/my_init.d/uid.sh /etc/my_init.d/uid.sh



#
# Cleanup & optimize
#

RUN find /var/log -type f -delete && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*