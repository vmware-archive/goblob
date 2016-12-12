FROM ubuntu:16.04
MAINTAINER Caleb Washburn <cwashburn@pivotal.io>

RUN apt-get update && apt-get install -y openssh-server
RUN mkdir /var/run/sshd
RUN echo 'root:screencast' | chpasswd
RUN sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN useradd -c "vcap" -m -s /bin/bash vcap
RUN echo 'vcap:password' | chpasswd

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile

RUN mkdir -p /var/vcap/store/shared

COPY blobstore/fixtures /var/vcap/store/shared

RUN chown vcap:vcap -R /var/vcap

EXPOSE 2222
CMD ["/usr/sbin/sshd", "-D", "-p", "2222"]
